package effects

/*
 * Data structure representing a tremolo effect.
 */
type tremolo struct {
	unitStruct
	attenuated   bool
	inStateSince uint32
}

/*
 * Tremolo audio processing.
 */
func (this *tremolo) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	frequency, _ := this.getNumericValue("frequency")
	phase, _ := this.getNumericValue("phase")
	depth, _ := this.getNumericValue("depth")
	this.mutex.RUnlock()
	sampleRateFloat := float64(sampleRate)
	frequencyFloat := float64(frequency)
	frequencyValue := 0.1 * frequencyFloat
	periodLengthFloat := sampleRateFloat / frequencyValue
	periodLength := uint32(periodLengthFloat)
	phaseFloat := float64(phase)
	phaseValue := 0.01 * phaseFloat
	samplesUnattenuatedFloat := periodLengthFloat * phaseValue
	samplesUnattenuated := uint32(samplesUnattenuatedFloat)
	samplesAttenuated := periodLength - samplesUnattenuated
	fac := decibelsToFactor(depth)
	attenuated := this.attenuated
	inStateSince := this.inStateSince

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		result := sample

		/*
		 * Perform state transitions.
		 */
		if attenuated && (inStateSince >= samplesAttenuated) {
			attenuated = false
			inStateSince = 0
		} else if !attenuated && (inStateSince >= samplesUnattenuated) {
			attenuated = true
			inStateSince = 0
		}

		/*
		 * Check if signal should be attenuated.
		 */
		if attenuated {
			result *= fac
		}

		out[i] = result
		inStateSince++
	}

	this.attenuated = attenuated
	this.inStateSince = inStateSince
}

/*
 * Create a tremolo effects unit.
 */
func createTremolo() Unit {

	/*
	 * Create effects unit.
	 */
	u := tremolo{
		unitStruct: unitStruct{
			unitType: UNIT_TREMOLO,
			params: []Parameter{
				Parameter{
					Name:               "frequency",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            10,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "phase",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       50,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "depth",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -10,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
