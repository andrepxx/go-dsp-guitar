package effects

import (
	"math"
)

/*
 * Data structure representing an auto wah effect.
 */
type autowah struct {
	unitStruct
	envelope            float64
	highpassCapVoltages [NUM_FILTERS]float64
	lowpassCapVoltages  [NUM_FILTERS]float64
}

/*
 * Auto wah audio processing.
 */
func (this *autowah) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	follow, _ := this.getDiscreteValue("follow")
	levelA, _ := this.getNumericValue("level_1")
	levelB, _ := this.getNumericValue("level_2")
	frequencyA, _ := this.getNumericValue("frequency_1")
	frequencyB, _ := this.getNumericValue("frequency_2")
	this.mutex.RUnlock()

	/*
	 * If the first level is higher than the second, swap both levels and frequencies around.
	 */
	if levelA > levelB {
		levelA, levelB = levelB, levelA
		frequencyA, frequencyB = frequencyB, frequencyA
	}

	levelAFloat := float64(levelA)
	levelBFloat := float64(levelB)
	frequencyAFloat := float64(frequencyA)
	frequencyBFloat := float64(frequencyB)
	frequencySlope := (frequencyBFloat - frequencyAFloat) / (levelBFloat - levelAFloat)
	sampleRateFloat := float64(sampleRate)
	dischargePerSampleEnvelopeArg := -20.0 / sampleRateFloat
	dischargePerSampleEnvelopeInv := math.Exp(dischargePerSampleEnvelopeArg)
	dischargePerSampleEnvelope := 1.0 - dischargePerSampleEnvelopeInv
	envelope := this.envelope
	hcvs := this.highpassCapVoltages
	lcvs := this.lowpassCapVoltages
	gainCompensation := math.Pow(2.0, NUM_FILTERS)

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

		level := factorToDecibels(envelope)
		frequency := 0.0

		/*
		 * Calculate the current limit frequency of the filter as a piecewise
		 * linear function of the signal level.
		 */
		if level <= levelAFloat {
			frequency = frequencyAFloat
		} else if level >= levelBFloat {
			frequency = frequencyBFloat
		} else {
			excess := level - levelAFloat
			frequency = frequencyAFloat + (frequencySlope * excess)
		}

		arg := -frequency / sampleRateFloat
		dischargePerSampleInv := 1.0 - math.Exp(arg)
		lcv := sample

		/*
		 * Evaluate the response of all filters.
		 */
		for j := 0; j < NUM_FILTERS; j++ {
			hcv := hcvs[j]
			diff := lcv - hcv
			hcv += diff * dischargePerSampleInv
			hcvs[j] = hcv
			lcv = lcvs[j]
			diff -= lcv
			lcv += diff * dischargePerSampleInv
			lcvs[j] = lcv
		}

		pre := gainCompensation * lcv

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
	this.highpassCapVoltages = hcvs
	this.lowpassCapVoltages = lcvs
}

/*
 * Create an auto-wah effects unit.
 */
func createAutoWah() Unit {

	/*
	 * Create effects unit.
	 */
	u := autowah{
		unitStruct: unitStruct{
			unitType: UNIT_AUTOWAH,
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
					Name:               "level_1",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -40,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_2",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -10,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "frequency_1",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            20000,
					NumericValue:       300,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "frequency_2",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            20000,
					NumericValue:       6000,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
