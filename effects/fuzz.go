package effects

import (
	"github.com/andrepxx/go-dsp-guitar/oversampling"
	"math"
)

/*
 * Data structure representing a fuzz effect.
 */
type fuzz struct {
	unitStruct
	bufferIn                 []float64
	bufferOut                []float64
	oversamplerTwo           oversampling.OversamplerDecimator
	oversamplerFour          oversampling.OversamplerDecimator
	envelope                 float64
	couplingCapacitorVoltage float64
}

/*
 * Internal (oversampled) fuzz audio processing.
 */
func (this *fuzz) processOversampled(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	follow, _ := this.getDiscreteValue("follow")
	bias, _ := this.getNumericValue("bias")
	boost, _ := this.getNumericValue("boost")
	gain, _ := this.getNumericValue("gain")
	fuzz, _ := this.getNumericValue("fuzz")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	biasFloat := float64(bias)
	biasFactor := 0.01 * biasFloat
	gainFactor := decibelsToFactor(boost + gain)
	fuzzFloat := float64(fuzz)
	fuzzFactor := 0.01 * fuzzFloat
	fuzzFactorInv := 1.0 - fuzzFactor
	levelFactor := decibelsToFactor(level)
	envelope := this.envelope
	couplingCapacitorVoltage := this.couplingCapacitorVoltage
	sampleRateFloat := float64(sampleRate)
	dischargePerSampleArg := -20.0 / sampleRateFloat
	dischargePerSampleInv := math.Exp(dischargePerSampleArg)
	dischargePerSample := 1.0 - dischargePerSampleInv

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		sampleAbs := math.Abs(sample)

		/*
		 * Follow either level or envelope.
		 */
		switch follow {
		case "envelope":
			envelope *= dischargePerSampleInv

			/*
			 * If the absolute value of the current sample exceeds the
			 * current envelope value, make it the new envelope value.
			 */
			if sampleAbs > envelope {
				envelope = sampleAbs
			}

		case "level":
			diff := sampleAbs - envelope
			envelope += diff * dischargePerSample
		default:
			envelope = 1.0
		}

		biasVoltage := biasFactor * envelope
		pre := gainFactor * (sample - biasVoltage)

		/*
		 * Clip the waveform.
		 */
		if pre < -1.0 {
			pre = -1.0
		} else if pre > 1.0 {
			pre = 1.0
		}

		fuzzFraction := fuzzFactor * pre
		cleanFraction := fuzzFactorInv * sample
		pre = fuzzFraction + cleanFraction
		diff := pre - couplingCapacitorVoltage
		couplingCapacitorVoltage += diff * dischargePerSample
		pre -= couplingCapacitorVoltage

		/*
		 * Limit the signal to the appropriate range.
		 */
		if pre < -1.0 {
			pre = -1.0
		} else if pre > 1.0 {
			pre = 1.0
		}

		out[i] = levelFactor * pre
	}

	this.envelope = envelope
	this.couplingCapacitorVoltage = couplingCapacitorVoltage
}

/*
 * Fuzz audio processing.
 */
func (this *fuzz) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	oversampling, _ := this.getDiscreteValue("oversampling")
	this.mutex.RUnlock()
	factor := 1

	/*
	 * Enable two- or four-times oversampling.
	 */
	switch oversampling {
	case "2":
		factor = 2
	case "4":
		factor = 4
	}

	/*
	 * Check if we require oversampling.
	 */
	if factor > 1 {
		numSamples := factor * len(in)
		bufferIn := this.bufferIn

		/*
		 * Ensure that the oversampled input buffer has sufficient
		 * size.
		 */
		if len(bufferIn) != numSamples {
			bufferIn = make([]float64, numSamples)
			this.bufferIn = bufferIn
		}

		bufferOut := this.bufferOut

		/*
		 * Ensure that the oversampled output buffer has sufficient
		 * size.
		 */
		if len(bufferOut) != numSamples {
			bufferOut = make([]float64, numSamples)
			this.bufferOut = bufferOut
		}

		oversampler := this.oversamplerTwo

		if factor == 4 {
			oversampler = this.oversamplerFour
		}

		oversampler.Oversample(in, bufferIn)
		factor32 := uint32(factor)
		oversampledRate := factor32 * sampleRate
		this.processOversampled(bufferIn, bufferOut, oversampledRate)
		oversampler.Decimate(bufferOut, out)
	} else {
		this.processOversampled(in, out, sampleRate)
	}

}

/*
 * Create a fuzz effects unit.
 */
func createFuzz() Unit {
	oversamplerTwo := oversampling.CreateOversamplerDecimator(2)
	oversamplerFour := oversampling.CreateOversamplerDecimator(4)

	/*
	 * Create effects unit.
	 */
	u := fuzz{
		unitStruct: unitStruct{
			unitType: UNIT_FUZZ,
			params: []Parameter{
				Parameter{
					Name:               "follow",
					Type:               PARAMETER_TYPE_DISCRETE,
					PhysicalUnit:       "",
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 1,
					DiscreteValues: []string{
						"envelope",
						"level",
					},
				},
				Parameter{
					Name:               "bias",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "%",
					Minimum:            -100,
					Maximum:            100,
					NumericValue:       50,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "boost",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            0,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "gain",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -30,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "fuzz",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "%",
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "oversampling",
					Type:               PARAMETER_TYPE_DISCRETE,
					PhysicalUnit:       "",
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 0,
					DiscreteValues: []string{
						"- NONE -",
						"2",
						"4",
					},
				},
			},
		},
		oversamplerTwo:  oversamplerTwo,
		oversamplerFour: oversamplerFour,
	}

	return &u
}
