package effects

import (
	"math"
)

/*
 * Data structure representing a ring modulator effect.
 */
type ringModulator struct {
	unitStruct
	phase float64
}

/*
 * Ring modulator audio processing.
 */
func (this *ringModulator) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	frequency, _ := this.getNumericValue("frequency")
	this.mutex.RUnlock()
	phase := this.phase
	sampleRateFloat := float64(sampleRate)
	frequencyFloat := float64(frequency)
	angularFrequency := MATH_TWO_PI * frequencyFloat
	phaseFraction := angularFrequency / sampleRateFloat

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		iFloat := float64(i)
		phaseOffset := iFloat * phaseFraction
		phaseUpdate := phase + phaseOffset
		currentPhase := math.Mod(phaseUpdate, MATH_TWO_PI)
		carrierWave := math.Sin(currentPhase)
		out[i] = carrierWave * sample
	}

	n := len(in)
	nFloat := float64(n)
	phaseOffset := nFloat * phaseFraction
	phaseUpdate := phase + phaseOffset
	this.phase = math.Mod(phaseUpdate, MATH_TWO_PI)
}

/*
 * Create a ring modulator effects unit.
 */
func createRingModulator() Unit {

	/*
	 * Create effects unit.
	 */
	u := ringModulator{
		unitStruct: unitStruct{
			unitType: UNIT_RINGMODULATOR,
			params: []Parameter{
				Parameter{
					Name:               "frequency",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
