package effects

import (
	"github.com/andrepxx/go-dsp-guitar/oversampling"
	"math"
)

/*
 * Data structure representing an overdrive effect.
 */
type overdrive struct {
	unitStruct
	bufferIn        []float64
	bufferOut       []float64
	oversamplerTwo  oversampling.OversamplerDecimator
	oversamplerFour oversampling.OversamplerDecimator
}

/*
 * Internal (oversampled) overdrive audio processing.
 */
func (this *overdrive) processOversampled(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	boost, _ := this.getNumericValue("boost")
	gain, _ := this.getNumericValue("gain")
	drive, _ := this.getNumericValue("drive")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	totalGain := boost + gain
	gainFactor := decibelsToFactor(totalGain)
	driveFloat := float64(drive)
	driveFactor := 0.01 * driveFloat
	cleanFactor := 1.0 - driveFactor
	levelFactor := decibelsToFactor(level)

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		arg := -gainFactor * sample
		x := math.Exp(arg)
		dist := (2.0 / (1.0 + x)) - 1.0
		mix := (driveFactor * dist) + (cleanFactor * sample)
		out[i] = levelFactor * mix
	}

}

/*
 * Overdrive audio processing.
 */
func (this *overdrive) Process(in []float64, out []float64, sampleRate uint32) {
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
 * Create an overdrive effects unit.
 */
func createOverdrive() Unit {
	oversamplerTwo := oversampling.CreateOversamplerDecimator(2)
	oversamplerFour := oversampling.CreateOversamplerDecimator(4)

	/*
	 * Create effects unit.
	 */
	u := overdrive{
		unitStruct: unitStruct{
			unitType: UNIT_OVERDRIVE,
			params: []Parameter{
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
					Name:               "drive",
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
