package effects

import (
	"math"
)

/*
 * Data structure representing a phaser effect.
 */
type phaser struct {
	unitStruct
	buffer        []float64
	previousPhase float64
}

/*
 * Phaser audio processing.
 */
func (this *phaser) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	depth, _ := this.getNumericValue("depth")
	speed, _ := this.getNumericValue("speed")
	phase, _ := this.getNumericValue("phase")
	this.mutex.RUnlock()
	depthFloat := float64(depth)
	depthValue := 0.001 * depthFloat

	/*
	 * Limit depth to [0.0; 1.0].
	 */
	if depthValue < 0.0 {
		depthValue = 0.0
	} else if depthValue > 1.0 {
		depthValue = 1.0
	}

	speedFloat := float64(speed)
	angularSpeed := MATH_TWO_PI_HUNDREDTH * speedFloat
	phaseFloat := float64(phase)
	phaseFloatRadians := MATH_DEGREE_TO_RADIANS * phaseFloat
	phaseFac := 0.5 * math.Sin(phaseFloatRadians)
	phaseFacInv := 1.0 - math.Abs(phaseFac)
	sampleRateFloat := float64(sampleRate)
	sampleRateFloatInv := 1.0 / sampleRateFloat
	maxDelaySamplesFloat := math.Floor((0.002 * sampleRateFloat) + 0.5)
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
	 * Mix the straight output with the delayed signal.
	 */
	for i, sample := range in {
		iFloat := float64(i)
		time := iFloat * sampleRateFloatInv
		phaseChange := angularSpeed * time
		phaseChanged := previousPhase + phaseChange
		phase := math.Mod(phaseChanged, MATH_TWO_PI)
		offset := depthValue * math.Sin(phase)
		currentDelayTime := 0.001 * (depthValue + offset)
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
		delayedSample := (weightEarly * delayedSampleEarly) + (weightLate * delayedSampleLate)
		out[i] = (phaseFacInv * sample) + (phaseFac * delayedSample)
	}

	bufferSizeFloat := float64(bufferSize)
	duration := bufferSizeFloat * sampleRateFloatInv
	phaseIncrement := angularSpeed * duration
	this.previousPhase = math.Mod(previousPhase+phaseIncrement, MATH_TWO_PI)
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
 * Create a phaser effects unit.
 */
func createPhaser() Unit {

	/*
	 * Create effects unit.
	 */
	u := phaser{
		unitStruct: unitStruct{
			unitType: UNIT_PHASER,
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
					NumericValue:       10,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "phase",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -90,
					Maximum:            90,
					NumericValue:       45,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
