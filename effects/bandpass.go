package effects

import (
	"math"
	"strconv"
)

/*
 * Data structure representing a bandpass effect.
 */
type bandpass struct {
	unitStruct
	highpassCapVoltages []float64
	lowpassCapVoltages  []float64
}

/*
 * Bandpass audio processing.
 */
func (this *bandpass) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	filterOrderString, _ := this.getDiscreteValue("filter_order")
	frequencyA, _ := this.getNumericValue("frequency_1")
	frequencyB, _ := this.getNumericValue("frequency_2")
	this.mutex.RUnlock()
	filterOrder, _ := strconv.ParseUint(filterOrderString, 10, 32)
	halfOrderUint := filterOrder >> 1
	halfOrder := int(halfOrderUint)

	/*
	 * If the first frequency is higher than the second, swap them around.
	 */
	if frequencyA > frequencyB {
		frequencyA, frequencyB = frequencyB, frequencyA
	}

	/*
	 * Allocate storage for highpass capacitor voltages if needed.
	 */
	if len(this.highpassCapVoltages) != halfOrder {
		this.highpassCapVoltages = make([]float64, halfOrder)
	}

	/*
	 * Allocate storage for lowpass capacitor voltages if needed.
	 */
	if len(this.lowpassCapVoltages) != halfOrder {
		this.lowpassCapVoltages = make([]float64, halfOrder)
	}

	sampleRateFloat := float64(sampleRate)
	minusTwoPiOverSampleRate := -MATH_TWO_PI / sampleRateFloat
	frequencyAFloat := float64(frequencyA)
	frequencyBFloat := float64(frequencyB)
	dischargePerSampleHPArg := minusTwoPiOverSampleRate * frequencyAFloat
	dischargePerSampleHP := math.Exp(dischargePerSampleHPArg)
	dischargePerSampleHPInv := 1.0 - dischargePerSampleHP
	dischargePerSampleLPArg := minusTwoPiOverSampleRate * frequencyBFloat
	dischargePerSampleLP := math.Exp(dischargePerSampleLPArg)
	dischargePerSampleLPInv := 1.0 - dischargePerSampleLP

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		pre := sample

		/*
		 * Filter as many times as required by the filter order.
		 */
		for j := 0; j < halfOrder; j++ {
			hcv := this.highpassCapVoltages[j]
			diff := pre - hcv
			hcv += diff * dischargePerSampleHPInv
			this.highpassCapVoltages[j] = hcv
			lcv := this.lowpassCapVoltages[j]
			diff -= lcv
			iv := lcv
			lcv += diff * dischargePerSampleLPInv
			this.lowpassCapVoltages[j] = lcv

			/*
			 * Limit the output signal to the appropriate range.
			 */
			if iv < -1.0 {
				pre = -1.0
			} else if iv > 1.0 {
				pre = 1.0
			} else {
				pre = iv
			}

		}

		out[i] = pre
	}

}

/*
 * Create a bandpass effects unit.
 */
func createBandpass() Unit {

	/*
	 * Create effects unit.
	 */
	u := bandpass{
		unitStruct: unitStruct{
			unitType: UNIT_BANDPASS,
			params: []Parameter{
				Parameter{
					Name:               "filter_order",
					Type:               PARAMETER_TYPE_DISCRETE,
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 0,
					DiscreteValues: []string{
						"2",
						"4",
						"6",
						"8",
					},
				},
				Parameter{
					Name:               "frequency_1",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            20000,
					NumericValue:       300,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "frequency_2",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            20000,
					NumericValue:       3000,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
