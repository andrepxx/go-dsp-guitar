package effects

import (
	"math"
)

/*
 * Data structure representing an octaver effect.
 */
type octaver struct {
	unitStruct
	previousPolarity         float64
	octaveRegister           uint32
	envelope                 float64
	couplingCapacitorVoltage float64
}

/*
 * Octaver audio processing.
 */
func (this *octaver) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	follow, _ := this.getDiscreteValue("follow")
	levelOctaveUp, _ := this.getNumericValue("level_octave_up")
	levelClean, _ := this.getNumericValue("level_clean")
	levelDist, _ := this.getNumericValue("level_dist")
	levelOctaveDownFirst, _ := this.getNumericValue("level_octave_down_first")
	levelOctaveDownSecond, _ := this.getNumericValue("level_octave_down_second")
	levelHysteresis, _ := this.getNumericValue("level_hysteresis")
	this.mutex.RUnlock()
	facOctaveUp := decibelsToFactor(levelOctaveUp)
	facClean := decibelsToFactor(levelClean)
	facDist := decibelsToFactor(levelDist)
	facOctaveDownFirst := decibelsToFactor(levelOctaveDownFirst)
	facOctaveDownSecond := decibelsToFactor(levelOctaveDownSecond)
	facHysteresis := decibelsToFactor(levelHysteresis)
	previousPolarity := this.previousPolarity
	octaveRegister := this.octaveRegister
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

		square := sample * sample
		sign := signFloat(sample)
		hysteresis := envelope * facHysteresis

		/*
		 * If signal changes polarity and is above the hysteresis, increment
		 * two-bit octave register.
		 */
		if (sign != 0.0) && (sign != previousPolarity) && (sampleAbs > hysteresis) {
			octaveRegister = (octaveRegister + 1) & 0x7
			previousPolarity = sign
		}

		firstDown := float64(1.0)

		/*
		 * Invert polarity of first octave down, depending on the contents of the
		 * octave register.
		 */
		if (octaveRegister & 0x2) != 0 {
			firstDown = -1.0
		}

		secondDown := float64(1.0)

		/*
		 * Invert polarity of second octave down, depending on the contents of the
		 * octave register.
		 */
		if (octaveRegister & 0x4) != 0 {
			secondDown = -1.0
		}

		pre := facClean * sample

		/*
		 * Check that envelope is not too small.
		 */
		if envelope > 0.0001 {
			pre += facOctaveUp * (square / envelope)
		}

		pre += facDist * (sign * envelope)
		pre += facOctaveDownFirst * (firstDown * envelope)
		pre += facOctaveDownSecond * (secondDown * envelope)
		couplingCapacitorVoltage += (pre - couplingCapacitorVoltage) * dischargePerSample
		pre -= couplingCapacitorVoltage

		/*
		 * Limit the output signal to the appropriate range.
		 */
		if pre < -1.0 {
			out[i] = -1.0
		} else if pre > 1.0 {
			out[i] = 1.0
		} else {
			out[i] = pre
		}

	}

	this.previousPolarity = previousPolarity
	this.octaveRegister = octaveRegister
	this.envelope = envelope
	this.couplingCapacitorVoltage = couplingCapacitorVoltage
}

/*
 * Create an octaver effects unit.
 */
func createOctaver() Unit {

	/*
	 * Create effects unit.
	 */
	u := octaver{
		unitStruct: unitStruct{
			unitType: UNIT_OCTAVER,
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
					Name:               "level_octave_up",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_clean",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_dist",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_octave_down_first",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_octave_down_second",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_hysteresis",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
