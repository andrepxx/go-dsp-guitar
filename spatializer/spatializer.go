package spatializer

import (
	"fmt"
	"math"
	"sync"
)

/*
 * Mathematical constants.
 */
const (
	MATH_DEGREE_TO_RADIANS = math.Pi / 180.0
)

/*
 * Other global constants.
 */
const (
	DEFAULT_SAMPLE_RATE     = 96000
	EFFECTIVE_DISTANCE      = 0.215
	HALF_EFFECTIVE_DISTANCE = 0.5 * EFFECTIVE_DISTANCE
	GROUP_DELAY             = 6.3e-4
	OUTPUT_COUNT            = 2
)

/*
 * Interface type for a spatializer.
 */
type Spatializer interface {
	GetAzimuth(inputChannel int) (float64, error)
	GetDistance(inputChannel int) (float64, error)
	GetLevel(inputChannel int) (float64, error)
	GetInputCount() int
	GetOutputCount() int
	Process(inputBuffers [][]float64, auxInputBuffer []float64, outputBuffers [][]float64)
	SetAzimuth(inputChannel int, azimuth float64) error
	SetDistance(inputChannel int, distance float64) error
	SetLevel(inputChannel int, level float64) error
	SetSampleRate(rate uint32)
}

/*
 * Data structure representing the position of an audio source in space.
 */
type position struct {
	azimuth  float64
	distance float64
	level    float64
}

/*
 * Data structure representing a spatializer.
 */
type spatializerStruct struct {
	buffers    [][]float64
	inputCount int
	sampleRate uint32
	mutex      sync.RWMutex
	positions  []position
}

/*
 * Returns the azimuth value associated with a channel.
 */
func (this *spatializerStruct) GetAzimuth(inputChannel int) (float64, error) {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return 0.0, fmt.Errorf("Cannot get azimuth for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {
		this.mutex.RLock()
		az := this.positions[inputChannel].azimuth
		this.mutex.RUnlock()
		return az, nil
	}

}

/*
 * Returns the distance value associated with a channel.
 */
func (this *spatializerStruct) GetDistance(inputChannel int) (float64, error) {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return 0.0, fmt.Errorf("Cannot get distance for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {
		this.mutex.RLock()
		dist := this.positions[inputChannel].distance
		this.mutex.RUnlock()
		return dist, nil
	}

}

/*
 * Returns the level value associated with a channel.
 */
func (this *spatializerStruct) GetLevel(inputChannel int) (float64, error) {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return 0.0, fmt.Errorf("Cannot get level for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {
		this.mutex.RLock()
		level := this.positions[inputChannel].level
		this.mutex.RUnlock()
		return level, nil
	}

}

/*
 * Returns the number of input streams this spatializer processes.
 */
func (this *spatializerStruct) GetInputCount() int {
	return this.inputCount
}

/*
 * Returns the number of output streams this spatializer generates.
 */
func (this *spatializerStruct) GetOutputCount() int {
	return OUTPUT_COUNT
}

/*
 * Perform the spatializer audio processing.
 */
func (this *spatializerStruct) Process(inputBuffers [][]float64, auxInputBuffer []float64, outputBuffers [][]float64) {
	nInputBuffers := len(inputBuffers)
	nOutputBuffers := len(outputBuffers)

	/*
	 * Verify that we have as many input and output buffers as we expect.
	 */
	if (nInputBuffers == this.inputCount) && (nOutputBuffers == OUTPUT_COUNT) {
		sampleRateFloat := float64(this.sampleRate)

		/*
		 * Iterate over the output buffers.
		 */
		for _, buffer := range outputBuffers {

			/*
			 * Iterate over the current buffer and zero it.
			 */
			for i, _ := range buffer {
				buffer[i] = 0.0
			}

		}

		this.mutex.RLock()

		/*
		 * Iterate over the input channels.
		 */
		for i, inputBuffer := range inputBuffers {
			position := this.positions[i]
			azimuth := MATH_DEGREE_TO_RADIANS * position.azimuth
			distance := position.distance
			level := position.level
			currentBuffer := this.buffers[i]
			bufferSize := len(currentBuffer)
			sinAz, cosAz := math.Sincos(azimuth)
			xPosition := distance * sinAz
			yPosition := distance * cosAz
			xDistLeft := math.Abs(xPosition + (HALF_EFFECTIVE_DISTANCE))
			xDistRight := math.Abs(xPosition - (HALF_EFFECTIVE_DISTANCE))
			yDist := math.Abs(yPosition)
			yDistSquared := yDist * yDist
			xDistLeftSquared := xDistLeft * xDistLeft
			distLeft := math.Sqrt(xDistLeftSquared + yDistSquared)
			preLeft := 1.0 / distLeft

			/*
			 * Factors should not exceed unity.
			 */
			if preLeft > 1.0 {
				preLeft = 1.0
			}

			facLeft := level * preLeft
			xDistRightSquared := xDistRight * xDistRight
			distRight := math.Sqrt(xDistRightSquared + yDistSquared)
			preRight := 1.0 / distRight

			/*
			 * Factors should not exceed unity.
			 */
			if preRight > 1.0 {
				preRight = 1.0
			}

			facRight := level * preRight
			distDiff := distLeft - distRight
			delayTime := (GROUP_DELAY / EFFECTIVE_DISTANCE) * distDiff
			delayTimeAbs := math.Abs(delayTime)
			delaySamples := delayTimeAbs * sampleRateFloat
			delaySamplesEarly := math.Floor(delaySamples)
			delaySamplesEarlyInt := int(delaySamplesEarly)

			/*
			 * Ensure that the delay does not exceed the buffer size.
			 */
			if delaySamplesEarlyInt >= bufferSize {
				delaySamplesEarlyInt = bufferSize - 1
			}

			delaySamplesLate := math.Ceil(delaySamples)
			delaySamplesLateInt := int(delaySamplesLate)

			/*
			 * Ensure that the delay does not exceed the buffer size.
			 */
			if delaySamplesLateInt >= bufferSize {
				delaySamplesLateInt = bufferSize - 1
			}

			/*
			 * Process each sample.
			 */
			for j, currentSample := range inputBuffer {

				/*
				 * Perform simplified processing if delay time is exactly zero.
				 */
				if delayTime == 0.0 {
					outputBuffers[0][j] += facLeft * currentSample
					outputBuffers[1][j] += facRight * currentSample
				} else {
					delayedIdxEarly := j - delaySamplesEarlyInt
					delayedIdxLate := j - delaySamplesLateInt
					delayedSampleEarly := float64(0.0)
					delayedSampleLate := float64(0.0)

					/*
					 * Check whether the delayed sample can be found in the current input
					 * signal or the delay buffer.
					 */
					if delayedIdxEarly >= 0 {
						delayedSampleEarly = inputBuffer[delayedIdxEarly]
					} else {
						bufferPtr := bufferSize + delayedIdxEarly
						delayedSampleEarly = currentBuffer[bufferPtr]
					}

					/*
					 * Check whether the delayed sample can be found in the current input
					 * signal or the delay buffer.
					 */
					if delayedIdxLate >= 0 {
						delayedSampleLate = inputBuffer[delayedIdxLate]
					} else {
						bufferPtr := bufferSize + delayedIdxLate
						delayedSampleLate = currentBuffer[bufferPtr]
					}

					weightEarly := 1.0 - (delaySamples - delaySamplesEarly)
					weightLate := 1.0 - (delaySamplesLate - delaySamples)
					earlySample := weightEarly * delayedSampleEarly
					lateSample := weightLate * delayedSampleLate
					delayedSample := earlySample + lateSample

					/*
					 * When the delay time is positive, the left channel is delayed.
					 * When the delay time is negative, the right channel is delayed.
					 */
					if delayTime > 0.0 {
						outputBuffers[0][j] += facLeft * delayedSample
						outputBuffers[1][j] += facRight * inputBuffer[j]
					} else {
						outputBuffers[0][j] += facLeft * inputBuffer[j]
						outputBuffers[1][j] += facRight * delayedSample
					}

				}

			}

		}

		this.mutex.RUnlock()

		/*
		 * If we have an aux input, mix it in as well.
		 */
		if auxInputBuffer != nil {

			/*
			 * Process each sample.
			 */
			for j, sample := range auxInputBuffer {
				outputBuffers[0][j] += sample
				outputBuffers[1][j] += sample
			}

		}

		/*
		 * Iterate over the input channels again to update all buffers.
		 */
		for i, inputBuffer := range inputBuffers {
			numSamples := len(inputBuffer)
			currentBuffer := this.buffers[i]
			bufferSize := len(currentBuffer)
			boundary := bufferSize - numSamples

			/*
			 * Check whether our buffer is larger than the number of samples processed.
			 */
			if boundary >= 0 {
				copy(currentBuffer[0:boundary], currentBuffer[numSamples:bufferSize])
				copy(currentBuffer[boundary:bufferSize], inputBuffer)
			} else {
				copy(currentBuffer, inputBuffer[-boundary:numSamples])
			}

		}

	}

}

/*
 * Sets the azimuth of the audio source associated with a certain channel.
 */
func (this *spatializerStruct) SetAzimuth(inputChannel int, azimuth float64) error {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return fmt.Errorf("Cannot set azimuth for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {
		this.mutex.Lock()
		this.positions[inputChannel].azimuth = azimuth
		this.mutex.Unlock()
		return nil
	}

}

/*
 * Sets the distance of the audio source associated with a certain channel.
 */
func (this *spatializerStruct) SetDistance(inputChannel int, distance float64) error {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return fmt.Errorf("Cannot set distance for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {

		/*
		 * Verify that the distance is within limits.
		 */
		if distance < 0.0 || distance > 10.0 {
			return fmt.Errorf("%s", "Failed to set distance: Value must be within [0, 10].")
		} else {
			this.mutex.Lock()
			this.positions[inputChannel].distance = distance
			this.mutex.Unlock()
			return nil
		}

	}

}

/*
 * Sets the level of the audio source associated with a certain channel.
 */
func (this *spatializerStruct) SetLevel(inputChannel int, level float64) error {
	inputCount := this.inputCount

	/*
	 * Verify that the channel exists.
	 */
	if inputChannel > inputCount {
		return fmt.Errorf("Cannot set distance for channel %d: Only %d channels exist.", inputChannel, inputCount)
	} else {

		/*
		 * Verify that the distance is within limits.
		 */
		if level < 0.0 || level > 1.0 {
			return fmt.Errorf("%s", "Failed to set level: Value must be within [0, 1].")
		} else {
			this.mutex.Lock()
			this.positions[inputChannel].level = level
			this.mutex.Unlock()
			return nil
		}

	}

}

/*
 * Changes the sample rate and recreates all inner buffers.
 */
func (this *spatializerStruct) SetSampleRate(rate uint32) {
	sampleRateFloat := float64(rate)
	bufferSizeFloat := math.Ceil(sampleRateFloat * GROUP_DELAY)
	bufferSize := int(bufferSizeFloat)
	inputChannels := this.inputCount

	/*
	 * Create each inner buffer.
	 */
	for i := 0; i < inputChannels; i++ {
		this.buffers[i] = make([]float64, bufferSize)
	}

}

/*
 * Creates a new spatializer.
 */
func Create(inputChannels int) Spatializer {
	positions := make([]position, inputChannels)

	/*
	 * Set the levels to one by default.
	 */
	for i, _ := range positions {
		positions[i].level = 1.0
	}

	buffers := make([][]float64, inputChannels)
	sampleRateFloat := float64(DEFAULT_SAMPLE_RATE)
	bufferSizeFloat := math.Ceil(sampleRateFloat * GROUP_DELAY)
	bufferSize := int(bufferSizeFloat)

	/*
	 * Create each inner buffer.
	 */
	for i, _ := range buffers {
		buffers[i] = make([]float64, bufferSize)
	}

	/*
	 * Create the new spatializer.
	 */
	s := spatializerStruct{
		inputCount: inputChannels,
		sampleRate: DEFAULT_SAMPLE_RATE,
		positions:  positions,
		buffers:    buffers,
	}

	return &s
}
