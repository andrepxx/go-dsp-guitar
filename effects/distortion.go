package effects

/*
 * Data structure representing a distortion effect.
 */
type distortion struct {
	unitStruct
}

/*
 * Distortion audio processing.
 */
func (this *distortion) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	boost, _ := this.getNumericValue("boost")
	gain, _ := this.getNumericValue("gain")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	totalGain := boost + gain
	gainFactor := decibelsToFactor(totalGain)
	levelFactor := decibelsToFactor(level)

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		pre := gainFactor * sample

		/*
		 * Limit the output signal to the appropriate range.
		 */
		if pre < -1.0 {
			pre = -1.0
		} else if pre > 1.0 {
			pre = 1.0
		}

		out[i] = levelFactor * pre
	}

}

/*
 * Create a distortion effects unit.
 */
func createDistortion() Unit {

	/*
	 * Create effects unit.
	 */
	u := distortion{
		unitStruct: unitStruct{
			unitType: UNIT_DISTORTION,
			params: []Parameter{
				Parameter{
					Name:               "boost",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            30,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
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
