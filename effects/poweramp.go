package effects

import (
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/fft"
	"github.com/andrepxx/go-dsp-guitar/filter"
	"strconv"
)

/*
 * Data structure representing a power amplifier.
 */
type poweramp struct {
	unitStruct
	sampleRate       uint32
	impulseResponses filter.ImpulseResponses
	idCompiled       uint64
	idReceived       uint64
	currentFilter    filter.Filter
}

/*
 * Compile a new filter for this power amplifier.
 */
func (this *poweramp) compile(sampleRate uint32, id uint64) (filter.Filter, error) {
	irs := this.impulseResponses

	/*
	 * Verify that impulse responses are loaded.
	 */
	if irs == nil {
		return nil, fmt.Errorf("%s", "Could not compile filter: No impulse responses were loaded.")
	} else {
		targetOrder := uint32(0)
		targetOrderString, err := this.getDiscreteValue("filter_order")

		/*
		 * Set target filter order.
		 */
		if err == nil {
			targetOrder64, err := strconv.ParseUint(targetOrderString, 10, 32)

			/*
			 * Abort if error occured during parsing.
			 */
			if err != nil {
				return nil, fmt.Errorf("Could not parse filter target order: '%s'", targetOrderString)
			} else {
				targetOrder = uint32(targetOrder64)
			}

		}

		filters := make([]filter.Filter, NUM_FILTERS)

		/*
		 * Populate each filter.
		 */
		for i := 0; i < NUM_FILTERS; i++ {
			iInc := int64(i + 1)
			sIdxInc := strconv.FormatInt(iInc, 10)
			paramFilter := "filter_" + sIdxInc
			paramLevel := "level_" + sIdxInc
			name, errName := this.getDiscreteValue(paramFilter)
			level, errLevel := this.getNumericValue(paramLevel)

			/*
			 * Check if an error occured.
			 */
			if errName != nil || errLevel != nil {
				return nil, fmt.Errorf("Error parsing values for filter %d.", i)
			} else {

				/*
				 * Verify that this is actually a valid filter and not a dummy value.
				 */
				if name != STRING_NONE {
					fac := decibelsToFactor(level)
					flt := irs.CreateFilter(name, sampleRate)

					/*
					 * Check if filter was found.
					 */
					if flt == nil {
						return nil, fmt.Errorf("Failed to load filter '%s' for sample rate '%d'.", name, sampleRate)
					} else {

						/*
						 * Check if target order makes sense.
						 */
						if targetOrder > 0 {
							flt = flt.Reduce(targetOrder)
						}

						flt = flt.Normalize()
						flt = flt.Multiply(fac)
						filters[i] = flt
					}

				}

			}

		}

		fltComposite := filter.Empty(sampleRate)

		/*
		 * Add all other filters.
		 */
		for _, flt := range filters {
			fltComposite, err = fltComposite.Add(flt)

			/*
			 * Check for errors.
			 */
			if err != nil {
				return nil, fmt.Errorf("Failed to add filter: %s", err.Error())
			}

		}

		return fltComposite, nil
	}

}

/*
 * Sets a discrete parameter value for a power amplifier.
 */
func (this *poweramp) SetDiscreteValue(name string, value string) error {
	this.mutex.Lock()
	err := this.unitStruct.setDiscreteValue(name, value)

	/*
	 * If value was set, recompile filter.
	 */
	if err == nil {
		sr := this.sampleRate
		id := this.idCompiled + 1
		this.idCompiled = id
		flt, err := this.compile(sr, id)

		/*
		 * Check if filter was compiled.
		 */
		if err == nil {
			this.currentFilter = flt
		}

	}

	this.mutex.Unlock()
	return err
}

/*
 * Sets a numeric parameter value for a power amplifier.
 */
func (this *poweramp) SetNumericValue(name string, value int32) error {
	this.mutex.Lock()
	err := this.unitStruct.setNumericValue(name, value)

	/*
	 * If value was set, recompile filter.
	 */
	if err == nil {
		sr := this.sampleRate
		id := this.idCompiled + 1
		this.idCompiled = id
		flt, err := this.compile(sr, id)

		/*
		 * Check if filter was compiled.
		 */
		if err == nil {
			this.currentFilter = flt
		}

	}

	this.mutex.Unlock()
	return err
}

/*
 * Power amplifier audio processing.
 */
func (this *poweramp) Process(in []float64, out []float64, sampleRate uint32) {

	/*
	 * Check if sampling rate changed.
	 */
	if sampleRate != this.sampleRate {
		this.sampleRate = sampleRate
		sr := this.sampleRate
		id := this.idCompiled + 1
		this.idCompiled = id
		flt, err := this.compile(sr, id)

		/*
		 * Check if filter was compiled.
		 */
		if err == nil {
			this.currentFilter = flt
		}

	}

	flt := this.currentFilter

	/*
	 * If there is a filter, put the signal through it, otherwise write zeros to output.
	 */
	if flt != nil {
		flt.Process(in, out)
	} else {
		fft.ZeroFloat(out)
	}

}

/*
 * Populate the parameters of a power amplifier.
 */
func PreparePowerAmp(unit Unit, responses filter.ImpulseResponses) error {
	isPowerAmp := false

	/*
	 * Check if unit is a power amplifier.
	 */
	switch unit.(type) {
	case *poweramp:
		isPowerAmp = true
	}

	/*
	 * Check if the unit is a power amp.
	 */
	if !isPowerAmp {
		return fmt.Errorf("%s", "Cannot prepare power amp: Unit is not a power amp.")
	} else if responses == nil {
		return fmt.Errorf("%s", "Cannot prepare power amp: Impulse responses are nil.")
	} else {
		amp := unit.(*poweramp)
		names := responses.Names()
		params := amp.unitStruct.params

		/*
		 * Create name and gain values for each filter.
		 */
		for i := 0; i < NUM_FILTERS; i++ {
			iInc := int64(i + 1)
			sIdxInc := strconv.FormatInt(iInc, 10)
			namesExtended := []string{STRING_NONE}
			namesExtended = append(namesExtended, names...)

			/*
			 * Parameter for power amp type.
			 */
			paramType := Parameter{
				Name:               "filter_" + sIdxInc,
				Type:               PARAMETER_TYPE_DISCRETE,
				PhysicalUnit:       "",
				Minimum:            -1,
				Maximum:            -1,
				NumericValue:       -1,
				DiscreteValueIndex: 0,
				DiscreteValues:     namesExtended,
			}

			/*
			 * Parameter for power amp level.
			 */
			paramLevel := Parameter{
				Name:               "level_" + sIdxInc,
				Type:               PARAMETER_TYPE_NUMERIC,
				PhysicalUnit:       "dB",
				Minimum:            -60,
				Maximum:            0,
				NumericValue:       0,
				DiscreteValueIndex: -1,
				DiscreteValues:     nil,
			}

			params = append(params, paramType, paramLevel)
		}

		amp.unitStruct.params = params
		amp.impulseResponses = responses
		return nil
	}

}

/*
 * Create a power amp effects unit.
 */
func createPowerAmp() Unit {

	/*
	 * Create effects unit.
	 */
	u := poweramp{
		unitStruct: unitStruct{
			unitType: UNIT_POWERAMP,
			params: []Parameter{
				Parameter{
					Name:               "filter_order",
					Type:               PARAMETER_TYPE_DISCRETE,
					PhysicalUnit:       "",
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 14,
					DiscreteValues: []string{
						"64",
						"128",
						"256",
						"512",
						"1024",
						"2048",
						"4096",
						"8192",
						"16384",
						"32768",
						"65536",
						"131072",
						"262144",
						"524288",
						"1048576",
					},
				},
			},
		},
	}

	return &u
}
