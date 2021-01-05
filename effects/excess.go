package effects

import (
	"github.com/andrepxx/go-dsp-guitar/oversampling"
	"math"
)

/*
 * Data structure representing an excess effect.
 */
type excess struct {
	unitStruct
	bufferIn        []float64
	bufferOut       []float64
	oversamplerTwo  oversampling.OversamplerDecimator
	oversamplerFour oversampling.OversamplerDecimator
}

/*
 * Internal (oversampled) excess audio processing.
 */
func (this *excess) processOversampled(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	gain, _ := this.getNumericValue("gain")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	gainFactor := decibelsToFactor(gain)
	levelFactor := decibelsToFactor(level)

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		pre := gainFactor * sample
		absPre := math.Abs(pre)
		exceeded := absPre > 1.0
		negative := pre < 0.0
		absPreBiased := absPre + 1.0
		preBiasedFloor := math.Floor(absPreBiased)
		section := int32(0.5 * preBiasedFloor)
		sectionLSB := section % 2
		sectionOdd := sectionLSB != 0
		inverted := sectionOdd != (exceeded && negative)
		absPreInc := absPre + 1.0
		excess := math.Mod(absPreInc, 2.0)

		/*
		 * Check if range has been exceeded.
		 */
		if exceeded {

			/*
			 * Decide, whether we're in range, go from top to bottom or from bottom to top.
			 */
			if inverted {
				pre = 1.0 - excess
			} else {
				pre = excess - 1.0
			}

		}

		out[i] = levelFactor * pre
	}

}

/*
 * Excess audio processing.
 */
func (this *excess) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	oversampling, _ := this.getDiscreteValue("oversampling")
	this.mutex.RUnlock()
	factor := 1

	/*
	 * Enable two- or four-times oversampling.
	 */
	switch oversampling {
	case "2":
		factor = 2
	case "4":
		factor = 4
	}

	/*
	 * Check if we require oversampling.
	 */
	if factor > 1 {
		numSamples := factor * len(in)
		bufferIn := this.bufferIn

		/*
		 * Ensure that the oversampled input buffer has sufficient
		 * size.
		 */
		if len(bufferIn) != numSamples {
			bufferIn = make([]float64, numSamples)
			this.bufferIn = bufferIn
		}

		bufferOut := this.bufferOut

		/*
		 * Ensure that the oversampled output buffer has sufficient
		 * size.
		 */
		if len(bufferOut) != numSamples {
			bufferOut = make([]float64, numSamples)
			this.bufferOut = bufferOut
		}

		oversampler := this.oversamplerTwo

		/*
		 * Check oversampling factor.
		 */
		if factor == 4 {
			oversampler = this.oversamplerFour
		}

		oversampler.Oversample(in, bufferIn)
		factor32 := uint32(factor)
		oversampledRate := factor32 * sampleRate
		this.processOversampled(bufferIn, bufferOut, oversampledRate)
		oversampler.Decimate(bufferOut, out)
	} else {
		this.processOversampled(in, out, sampleRate)
	}

}

/*
 * Create an excess effects unit.
 */
func createExcess() Unit {
	oversamplerTwo := oversampling.CreateOversamplerDecimator(2)
	oversamplerFour := oversampling.CreateOversamplerDecimator(4)

	/*
	 * Create effects unit.
	 */
	u := excess{
		unitStruct: unitStruct{
			unitType: UNIT_EXCESS,
			params: []Parameter{
				Parameter{
					Name:               "gain",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -30,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level",
					Type:               PARAMETER_TYPE_NUMERIC,
					PhysicalUnit:       "dB",
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "oversampling",
					Type:               PARAMETER_TYPE_DISCRETE,
					PhysicalUnit:       "",
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 0,
					DiscreteValues: []string{
						"- NONE -",
						"2",
						"4",
					},
				},
			},
		},
		oversamplerTwo:  oversamplerTwo,
		oversamplerFour: oversamplerFour,
	}

	return &u
}
