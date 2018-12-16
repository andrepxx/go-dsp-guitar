package effects

import (
	"math"
)

/*
 * Data structure representing a delay effect.
 */
type delay struct {
	unitStruct
	buffer []float64
}

/*
 * Delay audio processing.
 */
func (this *delay) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	delayTime, _ := this.getNumericValue("delay_time")
	feedback, _ := this.getNumericValue("feedback")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	delayTimeFloat := float64(delayTime)
	delayTimeSeconds := 0.001 * delayTimeFloat
	sampleRateFloat := float64(sampleRate)
	delaySamplesFloat := math.Floor((delayTimeSeconds * sampleRateFloat) + 0.5)
	delaySamples := int(delaySamplesFloat)
	feedbackFactor := decibelsToFactor(feedback)
	levelFactor := decibelsToFactor(level)
	bufferSize := len(this.buffer)

	/*
	 * Make sure the buffer has the appropriate size.
	 */
	if bufferSize != delaySamples {
		this.buffer = make([]float64, delaySamples)
		bufferSize = delaySamples
	}

	/*
	 * Mix the straight output with the delayed signal.
	 */
	for i, sample := range in {
		delayedIdx := i - delaySamples
		delayedSample := float64(0.0)

		/*
		 * Check whether the delayed sample can be found in the current input
		 * signal or the delay buffer.
		 */
		if delayedIdx >= 0 {
			delayedSample = in[delayedIdx]
		} else {
			bufferPtr := bufferSize + delayedIdx
			delayedSample = this.buffer[bufferPtr]
		}

		pre := levelFactor * (sample + (feedbackFactor * delayedSample))

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
 * Create a delay effects unit.
 */
func createDelay() Unit {

	/*
	 * Create effects unit.
	 */
	u := delay{
		unitStruct: unitStruct{
			unitType: UNIT_DELAY,
			params: []Parameter{
				Parameter{
					Name:               "delay_time",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            1000,
					NumericValue:       200,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "feedback",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -5,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       -5,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
