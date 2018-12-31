package effects

import (
	"math"
)

/*
 * Data structure representing an allpass filter.
 */
type reverbAllpass struct {
	buffer   []float64
	ptr      int
	feedback float64
}

/*
 * Data structure representing a tapped delay line.
 */
type reverbDelayLine struct {
	buffer  []float64
	indices []uint32
	factors []float64
}

/*
 * Data structure representing a reverb effect.
 */
type reverb struct {
	unitStruct
	allpasses       []*reverbAllpass
	delayLine       *reverbDelayLine
	frontBuffer     []float64
	backBuffer      []float64
	delayLineBuffer []float64
	sampleRate      uint32
}

/*
 * Allpass filter processing.
 */
func (this *reverbAllpass) process(in []float64, out []float64, sampleRate uint32) {
	buf := this.buffer
	bufSize := len(buf)
	ptrWrite := this.ptr
	feedback := this.feedback

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		ptrRead := (ptrWrite + 1) % bufSize
		delayedSample := buf[ptrRead]
		pre := sample - (feedback * delayedSample)
		buf[ptrWrite] = pre
		out[i] = (feedback * pre) + delayedSample
		ptrWrite = ptrRead
	}

	this.ptr = ptrWrite
}

/*
 * Tapped delay line processing.
 */
func (this *reverbDelayLine) process(in []float64, out []float64, sampleRate uint32) {
	buffer := this.buffer
	bufferSize := len(buffer)
	indices := this.indices
	factors := this.factors

	/*
	 * Calculate each output sample.
	 */
	for i := range out {
		pre := 0.0

		/*
		 * Iterate over each tap and sum them all up.
		 */
		for j, offset := range indices {
			offsetInt := int(offset)
			idx := i - offsetInt
			currentSample := 0.0

			/*
			 * If index is positive, take the sample from the input
			 * buffer, otherwise take it from the internal buffer.
			 */
			if idx >= 0 {
				currentSample = in[idx]
			} else if idx >= -bufferSize {
				bufIdx := bufferSize + idx
				currentSample = buffer[bufIdx]
			}

			fac := factors[j]
			pre += fac * currentSample
		}

		out[i] = pre
	}

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
 * Creates an allpass filter used for reverberation.
 */
func (this *reverb) createAllpass(samplesDelay int, feedback float64) *reverbAllpass {
	buf := make([]float64, samplesDelay)

	/*
	 * Create allpass filter.
	 */
	res := reverbAllpass{
		buffer:   buf,
		ptr:      0,
		feedback: feedback,
	}

	return &res
}

/*
 * Creates a tapped delay line used for reverberation.
 */
func (this *reverb) createDelayLine(indices []uint32, factors []float64) *reverbDelayLine {
	maxIndex := uint32(0)

	/*
	 * Find the maximum index.
	 */
	for _, index := range indices {

		/*
		 * Check if we found a larger index.
		 */
		if index > maxIndex {
			maxIndex = index
		}

	}

	buf := make([]float64, maxIndex)
	numIndices := len(indices)
	indicesNew := make([]uint32, numIndices)
	copy(indicesNew, indices)
	numFactors := len(factors)
	factorsNew := make([]float64, numFactors)
	copy(factorsNew, factors)

	/*
	 * Create tapped delay line.
	 */
	res := reverbDelayLine{
		buffer:  buf,
		indices: indicesNew,
		factors: factorsNew,
	}

	return &res
}

/*
 * Reverb audio processing.
 */
func (this *reverb) Process(in []float64, out []float64, sampleRate uint32) {
	nIn := len(in)
	nOut := len(out)

	/*
	 * Ensure that the input and output buffers are of equal size.
	 */
	if nIn != nOut {

		/*
		 * Write zeros to output buffer.
		 */
		for i, _ := range out {
			out[i] = 0.0
		}

	} else {
		this.mutex.RLock()
		mix, _ := this.getNumericValue("mix")
		this.mutex.RUnlock()
		mixFloat := float64(mix)
		wetFrac := 0.01 * mixFloat
		dryFrac := 1.0 - wetFrac
		sampleRateFloat := float64(sampleRate)

		/*
		 * If sample rate has changed, recreate all filter structures.
		 */
		if this.sampleRate != sampleRate {

			/*
			 * Delays for the allpass filters in seconds.
			 */
			allpassDelays := [...]float64{
				0.04204,
				0.01348,
				0.00452,
			}

			numAllpasses := len(allpassDelays)

			/*
			 * Make sure that the allpass array has the right size.
			 */
			if len(this.allpasses) != numAllpasses {
				this.allpasses = make([]*reverbAllpass, numAllpasses)
			}

			/*
			 * Create allpass filters.
			 */
			for i, delaySeconds := range allpassDelays {
				delaySamplesFloat := math.Round(delaySeconds * sampleRateFloat)
				delaySamples := int(delaySamplesFloat)
				allpass := this.createAllpass(delaySamples, 0.7)
				this.allpasses[i] = allpass
			}

			/*
			 * Time of delay line taps in seconds.
			 */
			delayLineTapsTime := []float64{
				0.19196,
				0.19996,
				0.21596,
				0.23204,
			}

			numTaps := len(delayLineTapsTime)
			delayLineTapsSamples := make([]uint32, numTaps)

			/*
			 * Calculate delay line taps in samples.
			 */
			for i, tapSeconds := range delayLineTapsTime {
				tapSamplesFloat := math.Round(tapSeconds * sampleRateFloat)
				tapSamples := uint32(tapSamplesFloat)
				delayLineTapsSamples[i] = tapSamples
			}

			/*
			 * Coefficients for delay line taps.
			 */
			delayLineTapsCoeffs := []float64{
				0.1855,
				0.18325,
				0.17875,
				0.17425,
			}

			delayLine := this.createDelayLine(delayLineTapsSamples, delayLineTapsCoeffs)
			this.delayLine = delayLine
			this.sampleRate = sampleRate
		}

		frontBuffer := this.frontBuffer
		backBuffer := this.backBuffer
		delayLineBuffer := this.delayLineBuffer

		/*
		 * Ensure that the front buffer has the correct size.
		 */
		if len(frontBuffer) != nIn {
			frontBuffer = make([]float64, nIn)
		}

		/*
		 * Ensure that the back buffer has the correct size.
		 */
		if len(backBuffer) != nIn {
			backBuffer = make([]float64, nIn)
		}

		/*
		 * Ensure that the delay line buffer has the correct size.
		 */
		if len(delayLineBuffer) != nIn {
			delayLineBuffer = make([]float64, nIn)
		}

		this.delayLine.process(in, delayLineBuffer, sampleRate)
		copy(frontBuffer, delayLineBuffer)

		/*
		 * Process the sound using the allpass filters.
		 */
		for _, allpass := range this.allpasses {
			allpass.process(frontBuffer, backBuffer, sampleRate)
			backBuffer, frontBuffer = frontBuffer, backBuffer
		}

		halfWetFrac := 0.5 * wetFrac

		/*
		 * Mix the dry and wet signal.
		 */
		for i, drySample := range in {
			delayedSample := delayLineBuffer[i]
			wetSample := frontBuffer[i]
			processedSampleSum := delayedSample + wetSample
			pre := (dryFrac * drySample) + (halfWetFrac * processedSampleSum)

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

		this.frontBuffer = frontBuffer
		this.backBuffer = backBuffer
		this.delayLineBuffer = delayLineBuffer
	}

}

/*
 * Create a reverb effects unit.
 */
func createReverb() Unit {

	/*
	 * Create effects unit.
	 */
	u := reverb{
		unitStruct: unitStruct{
			unitType: UNIT_REVERB,
			params: []Parameter{
				Parameter{
					Name:               "mix",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       50,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
