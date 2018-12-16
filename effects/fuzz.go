package effects

import (
	"math"
)

/*
 * Data structure representing a fuzz effect.
 */
type fuzz struct {
	unitStruct
	envelope                 float64
	couplingCapacitorVoltage float64
}

/*
 * Fuzz audio processing.
 */
func (this *fuzz) Process(in []float64, out []float64, sampleRate uint32) {
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
 * Create a fuzz effects unit.
 */
func createFuzz() Unit {

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
					Minimum:            -100,
					Maximum:            100,
					NumericValue:       50,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "boost",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "gain",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "fuzz",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
