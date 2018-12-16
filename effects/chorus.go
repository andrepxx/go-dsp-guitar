package effects

import (
	"math"
)

/*
 * Data structure representing a chorus effect.
 */
type chorus struct {
	unitStruct
	buffer        []float64
	previousPhase float64
}

/*
 * Chorus audio processing.
 */
func (this *chorus) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	depth, _ := this.getNumericValue("depth")
	speed, _ := this.getNumericValue("speed")
	this.mutex.RUnlock()
	depthFloat := 0.1 * float64(depth)

	/*
	 * Limit depth to [0.0; 10.0].
	 */
	if depthFloat < 0.0 {
		depthFloat = 0.0
	} else if depthFloat > 10.0 {
		depthFloat = 10.0
	}

	speedFloat := float64(speed)
	angularSpeed := MATH_PI_THOUSANDTH * speedFloat
	sampleRateFloat := float64(sampleRate)
	maxDelaySamplesFloat := math.Floor((0.05 * sampleRateFloat) + 0.5)
	maxDelaySamples := int(maxDelaySamplesFloat)
	bufferSize := len(this.buffer)
	previousPhase := this.previousPhase

	/*
	 * Make sure the buffer has the appropriate size.
	 */
	if bufferSize != maxDelaySamples {
		this.buffer = make([]float64, maxDelaySamples)
		bufferSize = maxDelaySamples
	}

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		iFloat := float64(i)
		time := iFloat / sampleRateFloat
		phaseChange := angularSpeed * time
		phaseChanged := previousPhase + phaseChange
		zeroPhase := math.Mod(phaseChanged, MATH_TWO_PI)
		effectedSample := 0.0

		/*
		 * Generate five sub-signals.
		 */
		for j := 0; j < 5; j++ {
			jFloat := float64(j)
			phaseOffset := MATH_TWO_PI_FIFTH * jFloat
			updatedPhase := zeroPhase + phaseOffset
			phase := math.Mod(updatedPhase, MATH_TWO_PI)
			offset := depthFloat * math.Sin(phase)
			currentDelayTime := 0.001 * (40.0 + offset)
			currentDelaySamples := currentDelayTime * sampleRateFloat
			currentDelaySamplesEarly := math.Floor(currentDelaySamples)
			currentDelaySamplesEarlyInt := int(currentDelaySamplesEarly)
			currentDelaySamplesLate := math.Ceil(currentDelaySamples)
			currentDelaySamplesLateInt := int(currentDelaySamplesLate)
			delayedIdxEarly := i - currentDelaySamplesEarlyInt
			delayedIdxLate := i - currentDelaySamplesLateInt
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
				delayedSampleEarly = this.buffer[bufferPtr]
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
			effectedSample += 0.2 * ((weightEarly * delayedSampleEarly) + (weightLate * delayedSampleLate))
		}

		out[i] = (0.5 * sample) + (0.5 * effectedSample)
	}

	bufferSizeFloat := float64(bufferSize)
	bufferTime := bufferSizeFloat / sampleRateFloat
	phaseChange := angularSpeed * bufferTime
	phaseChanged := previousPhase + phaseChange
	this.previousPhase = math.Mod(phaseChanged, MATH_TWO_PI)
	numSamples := len(in)
	boundary := bufferSize - numSamples

	/*
	 * Check whether our buffer is larger than the number of samples processed.
	 */
	if boundary >= 0 {
		copy(this.buffer[0:boundary], this.buffer[numSamples:bufferSize])
		copy(this.buffer[boundary:bufferSize], in)
	} else {
		copy(this.buffer, in[-boundary:numSamples])
	}

}

/*
 * Create a chorus effects unit.
 */
func createChorus() Unit {

	/*
	 * Create effects unit.
	 */
	u := chorus{
		unitStruct: unitStruct{
			unitType: UNIT_CHORUS,
			params: []Parameter{
				Parameter{
					Name:               "depth",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "speed",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            100,
					NumericValue:       30,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
