package effects

import (
	"math"
)

/*
 * Data structure representing an excess effect.
 */
type excess struct {
	unitStruct
}

/*
 * Excess audio processing.
 */
func (this *excess) Process(in []float64, out []float64, sampleRate uint32) {
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
 * Create an excess effects unit.
 */
func createExcess() Unit {

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
					Minimum:            -30,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "level",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
