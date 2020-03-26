package effects

import (
	"math"
)

const (
	NUM_HIGHPASS_FILTERS = 3
	NUM_LOWPASS_FILTERS  = 4
)

/*
 * Data structure representing a cabinet effect.
 */
type cabinet struct {
	unitStruct
	buffer                   []float64
	highpassCapVoltages      []float64
	highpassLimitFrequencies []float64
	lowpassCapVoltages       []float64
	lowpassLimitFrequencies  []float64
}

/*
 * Cabinet audio processing.
 */
func (this *cabinet) Process(in []float64, out []float64, sampleRate uint32) {
	highpassLimitFrequencies := this.highpassLimitFrequencies

	/*
	 * Check if limit frequencies for highpass filter are defined.
	 */
	if len(highpassLimitFrequencies) != NUM_HIGHPASS_FILTERS {

		/*
		 * Highpass filter limit frequencies.
		 */
		highpassLimitFrequencies = []float64{
			300.0,
			120.0,
			80.0,
		}

		this.highpassLimitFrequencies = highpassLimitFrequencies
	}

	highpassCapVoltages := this.highpassCapVoltages

	/*
	 * Make sure that there are as many capacitors in the HPF as required.
	 */
	if len(highpassCapVoltages) != NUM_HIGHPASS_FILTERS {
		highpassCapVoltages = make([]float64, NUM_HIGHPASS_FILTERS)
		this.highpassCapVoltages = highpassCapVoltages
	}

	lowpassLimitFrequencies := this.lowpassLimitFrequencies

	/*
	 * Check if limit frequencies for lowpass filter are defined.
	 */
	if len(lowpassLimitFrequencies) != NUM_LOWPASS_FILTERS {

		/*
		 * Lowpass filter limit frequencies.
		 */
		lowpassLimitFrequencies = []float64{
			3000.0,
			4000.0,
			5000.0,
			6000.0,
		}

		this.lowpassLimitFrequencies = lowpassLimitFrequencies
	}

	lowpassCapVoltages := this.lowpassCapVoltages

	/*
	 * Make sure that there are as many capacitors in the LPF as required.
	 */
	if len(lowpassCapVoltages) != NUM_LOWPASS_FILTERS {
		lowpassCapVoltages = make([]float64, NUM_LOWPASS_FILTERS)
		this.lowpassCapVoltages = lowpassCapVoltages
	}

	nIn := len(in)
	buffer := this.buffer

	/*
	 * Make sure that the buffer is of appropriate size.
	 */
	if len(buffer) != nIn {
		buffer = make([]float64, nIn)
		this.buffer = buffer
	}

	copy(buffer, in)
	sampleRateFloat := float64(sampleRate)
	minusTwoPiOverSampleRate := -MATH_TWO_PI / sampleRateFloat

	/*
	 * Process all highpass filters.
	 */
	for i, f := range highpassLimitFrequencies {
		hcv := this.highpassCapVoltages[i]
		dischargePerSampleArg := minusTwoPiOverSampleRate * f
		dischargePerSample := math.Exp(dischargePerSampleArg)
		dischargePerSampleInv := 1.0 - dischargePerSample

		/*
		 * Process each sample.
		 */
		for j, sample := range buffer {
			diff := sample - hcv
			buffer[j] = diff
			hcv += diff * dischargePerSampleInv
		}

		this.highpassCapVoltages[i] = hcv
	}

	/*
	 * Process all lowpass filters.
	 */
	for i, f := range lowpassLimitFrequencies {
		lcv := this.lowpassCapVoltages[i]
		dischargePerSampleArg := minusTwoPiOverSampleRate * f
		dischargePerSample := math.Exp(dischargePerSampleArg)
		dischargePerSampleInv := 1.0 - dischargePerSample

		/*
		 * Process each sample.
		 */
		for j, sample := range buffer {
			diff := sample - lcv
			buffer[j] = lcv
			lcv += diff * dischargePerSampleInv
		}

		this.lowpassCapVoltages[i] = lcv
	}

	/*
	 * Process each sample.
	 */
	for i, sample := range buffer {
		pre := sample

		/*
		 * Limit the output signal to the appropriate range.
		 */
		if pre < -1.0 {
			pre = -1.0
		} else if pre > 1.0 {
			pre = 1.0
		}

		out[i] = pre
	}

}

/*
 * Create a cabinet effects unit.
 */
func createCabinet() Unit {

	/*
	 * Create effects unit.
	 */
	u := cabinet{
		unitStruct: unitStruct{
			unitType: UNIT_CABINET,
			params: []Parameter{
				Parameter{
					Name:               "type",
					Type:               PARAMETER_TYPE_DISCRETE,
					PhysicalUnit:       "",
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 0,
					DiscreteValues: []string{
						"- DEFAULT -",
					},
				},
			},
		},
	}

	return &u
}
