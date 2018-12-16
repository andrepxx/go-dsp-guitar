package effects

import (
	"math"
)

/*
 * Data structure representing a tone stack effect.
 */
type toneStack struct {
	unitStruct
	highpassCapVoltages []float64
	lowpassCapVoltages  []float64
}

/*
 * Tone stack audio processing.
 */
func (this *toneStack) Process(in []float64, out []float64, sampleRate uint32) {
	frequencies := [...]float64{20.0, 300.0, 3000.0, 6000.0, 20000.0}
	facs := [...]float64{0.0, 0.0, 0.0, 0.0}
	names := [...]string{"low", "middle", "presence", "high"}
	numBands := len(facs)
	this.mutex.RLock()

	/*
	 * Read in levels and calculate factors.
	 */
	for i := 0; i < numBands; i++ {
		name := names[i]
		level, _ := this.getNumericValue(name)
		facs[i] = decibelsToFactor(level)
	}

	this.mutex.RUnlock()

	/*
	 * Allocate storage for highpass capacitor voltages if needed.
	 */
	if len(this.highpassCapVoltages) != numBands {
		this.highpassCapVoltages = make([]float64, numBands)
	}

	/*
	 * Allocate storage for lowpass capacitor voltages if needed.
	 */
	if len(this.lowpassCapVoltages) != numBands {
		this.lowpassCapVoltages = make([]float64, numBands)
	}

	sampleRateFloat := float64(sampleRate)
	minusTwoPiOverSampleRate := -MATH_TWO_PI / sampleRateFloat

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		sum := float64(0.0)

		/*
		 * Process each band and sum them all up.
		 */
		for j := 0; j < numBands; j++ {
			jInc := j + 1
			hcv := this.highpassCapVoltages[j]
			lcv := this.lowpassCapVoltages[j]
			frequencyA := frequencies[j]
			frequencyAFloat := float64(frequencyA)
			argHP := minusTwoPiOverSampleRate * frequencyAFloat
			dischargePerSampleHP := math.Exp(argHP)
			dischargePerSampleHPInv := 1.0 - dischargePerSampleHP
			frequencyB := frequencies[jInc]
			frequencyBFloat := float64(frequencyB)
			argLP := minusTwoPiOverSampleRate * frequencyBFloat
			dischargePerSampleLP := math.Exp(argLP)
			dischargePerSampleLPInv := 1.0 - dischargePerSampleLP
			diff := sample - hcv
			hcv += diff * dischargePerSampleHPInv
			diff -= lcv
			pre := lcv
			lcv += diff * dischargePerSampleLPInv
			this.highpassCapVoltages[j] = hcv
			this.lowpassCapVoltages[j] = lcv
			sum += facs[j] * pre
		}

		/*
		 * Limit the output signal to the appropriate range.
		 */
		if sum < -1.0 {
			out[i] = -1.0
		} else if sum > 1.0 {
			out[i] = 1.0
		} else {
			out[i] = sum
		}

	}

}

/*
 * Create a tone stack effects unit.
 */
func createToneStack() Unit {

	/*
	 * Create effects unit.
	 */
	u := toneStack{
		unitStruct: unitStruct{
			unitType: UNIT_TONESTACK,
			params: []Parameter{
				Parameter{
					Name:               "low",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "middle",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       -2,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "presence",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       -5,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "high",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -30,
					Maximum:            0,
					NumericValue:       -5,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
