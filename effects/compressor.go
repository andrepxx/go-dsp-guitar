package effects

import (
	"math"
)

/*
 * Data structure representing a compressor effect.
 */
type compressor struct {
	unitStruct
	envelope float64
}

/*
 * Compressor audio processing.
 */
func (this *compressor) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	follow, _ := this.getDiscreteValue("follow")
	gainLimit, _ := this.getNumericValue("gain_limit")
	targetLevel, _ := this.getNumericValue("target_level")
	this.mutex.RUnlock()
	gainLimitFac := decibelsToFactor(gainLimit)
	targetLevelFac := decibelsToFactor(targetLevel)
	sampleRateFloat := float64(sampleRate)
	dischargePerSampleEnvelopeArg := -20.0 / sampleRateFloat
	dischargePerSampleEnvelopeInv := math.Exp(dischargePerSampleEnvelopeArg)
	dischargePerSampleEnvelope := 1.0 - dischargePerSampleEnvelopeInv
	envelope := this.envelope

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
			envelope *= dischargePerSampleEnvelopeInv

			/*
			 * If the absolute value of the current sample exceeds the
			 * current envelope value, make it the new envelope value.
			 */
			if sampleAbs > envelope {
				envelope = sampleAbs
			}

		case "level":
			diff := sampleAbs - envelope
			envelope += diff * dischargePerSampleEnvelope
		default:
			envelope = 1.0
		}

		gain := targetLevelFac / envelope

		/*
		 * Limit maximum gain.
		 */
		if gain > gainLimitFac {
			gain = gainLimitFac
		}

		pre := gain * sample

		/*
		 * Limit the output signal to the appropriate range.
		 */
		if pre < -1.0 {
			pre = -1.0
		} else if pre > 1.0 {
			pre = 1.0
		}

		out[i] = pre
	}

	this.envelope = envelope
}

/*
 * Create a compressor effects unit.
 */
func createCompressor() Unit {

	/*
	 * Create effects unit.
	 */
	u := compressor{
		unitStruct: unitStruct{
			unitType: UNIT_COMPRESSOR,
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
					Name:               "gain_limit",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            30,
					NumericValue:       30,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "target_level",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
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
