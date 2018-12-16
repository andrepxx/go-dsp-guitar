package effects

import (
	"math"
)

/*
 * Data structure representing an overdrive effect.
 */
type overdrive struct {
	unitStruct
}

/*
 * Overdrive audio processing.
 */
func (this *overdrive) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	boost, _ := this.getNumericValue("boost")
	gain, _ := this.getNumericValue("gain")
	drive, _ := this.getNumericValue("drive")
	level, _ := this.getNumericValue("level")
	this.mutex.RUnlock()
	totalGain := boost + gain
	gainFactor := decibelsToFactor(totalGain)
	driveFloat := float64(drive)
	driveFactor := 0.01 * driveFloat
	cleanFactor := 1.0 - driveFactor
	levelFactor := decibelsToFactor(level)

	/*
	 * Process each sample.
	 */
	for i, sample := range in {
		arg := -gainFactor * sample
		x := math.Exp(arg)
		dist := (2.0 / (1.0 + x)) - 1.0
		mix := (driveFactor * dist) + (cleanFactor * sample)
		out[i] = levelFactor * mix
	}

}

/*
 * Create an overdrive effects unit.
 */
func createOverdrive() Unit {

	/*
	 * Create effects unit.
	 */
	u := overdrive{
		unitStruct: unitStruct{
			unitType: UNIT_OVERDRIVE,
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
					Name:               "drive",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
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
