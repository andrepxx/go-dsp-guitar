package effects

import (
	"math"
)

/*
 * Data structure representing an auto yoy effect.
 */
type autoyoy struct {
	unitStruct
	envelope float64
	buffer   []float64
}

/*
 * Auto wah audio processing.
 */
func (this *autoyoy) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	follow, _ := this.getDiscreteValue("follow")
	levelA, _ := this.getNumericValue("level_1")
	levelB, _ := this.getNumericValue("level_2")
	depth, _ := this.getNumericValue("depth")
	this.mutex.RUnlock()
	depthFloat := float64(depth)
	depthA := float64(0.0)
	depthB := 0.01 * depthFloat

	/*
	 * If the first level is higher than the second, swap both levels and depths around.
	 */
	if levelA > levelB {
		levelA, levelB = levelB, levelA
		depthA, depthB = depthB, depthA
	}

	levelAFloat := float64(levelA)
	levelBFloat := float64(levelB)
	depthSlope := (depthB - depthA) / (levelBFloat - levelAFloat)
	sampleRateFloat := float64(sampleRate)
	sampleRateFloatInv := 1.0 / sampleRateFloat
	dischargePerSampleEnvelopeArg := -20.0 * sampleRateFloatInv
	dischargePerSampleEnvelopeInv := math.Exp(dischargePerSampleEnvelopeArg)
	dischargePerSampleEnvelope := 1.0 - dischargePerSampleEnvelopeInv
	maxDelaySamplesFloat := math.Floor((0.01 * sampleRateFloat) + 0.5)
	maxDelaySamples := int(maxDelaySamplesFloat)
	buffer := this.buffer
	bufferSize := len(buffer)

	/*
	 * Make sure the buffer has the appropriate size.
	 */
	if bufferSize != maxDelaySamples {
		buffer = make([]float64, maxDelaySamples)
		this.buffer = buffer
		bufferSize = maxDelaySamples
	}

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

		level := factorToDecibels(envelope)
		delayFac := 0.0

		/*
		 * Calculate the current delay of the filter as a piecewise
		 * linear function of the signal level.
		 */
		if level <= levelAFloat {
			delayFac = depthA
		} else if level >= levelBFloat {
			delayFac = depthB
		} else {
			excess := level - levelAFloat
			delayFac = depthA + (depthSlope * excess)
		}

		currentDelayTime := 0.01 * delayFac
		currentDelaySamples := currentDelayTime * sampleRateFloat
		currentDelaySamplesEarly := math.Floor(currentDelaySamples)
		currentDelaySamplesLate := math.Ceil(currentDelaySamples)
		delayedIdxEarly := i - int(currentDelaySamplesEarly)
		delayedIdxLate := i - int(currentDelaySamplesLate)
		delayedSampleEarly := float64(0.0)
		delayedSampleLate := float64(0.0)

		/*
		 * Check whether the delayed sample can be found in the current input
		 * signal or the delay buffer.
		 */
		if delayedIdxEarly >= 0 {
			delayedSampleEarly = in[delayedIdxEarly]
		} else {
			bufferPtr := bufferSize + delayedIdxEarly
			delayedSampleEarly = buffer[bufferPtr]
		}

		/*
		 * Check whether the delayed sample can be found in the current input
		 * signal or the delay buffer.
		 */
		if delayedIdxLate >= 0 {
			delayedSampleLate = in[delayedIdxLate]
		} else {
			bufferPtr := bufferSize + delayedIdxLate
			delayedSampleLate = this.buffer[bufferPtr]
		}

		weightEarly := 1.0 - (currentDelaySamples - currentDelaySamplesEarly)
		weightLate := 1.0 - (currentDelaySamplesLate - currentDelaySamples)
		delayedSample := (weightEarly * delayedSampleEarly) + (weightLate * delayedSampleLate)
		out[i] = (0.5 * sample) + (0.5 * delayedSample)
	}

	this.envelope = envelope
	numSamples := len(in)
	boundary := bufferSize - numSamples

	/*
	 * Check whether our buffer is larger than the number of samples processed.
	 */
	if boundary >= 0 {
		copy(buffer[0:boundary], buffer[numSamples:bufferSize])
		copy(buffer[boundary:bufferSize], in)
	} else {
		copy(buffer, in[-boundary:numSamples])
	}
}

/*
 * Create an auto-yoy effects unit.
 */
func createAutoYoy() Unit {

	/*
	 * Create effects unit.
	 */
	u := autoyoy{
		unitStruct: unitStruct{
			unitType: UNIT_AUTOYOY,
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
					Name:               "level_1",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -40,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level_2",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -10,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "depth",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "%",
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
