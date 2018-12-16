package effects

import (
	"math"
)

/*
 * Data structure representing a noise gate effect.
 */
type noiseGate struct {
	unitStruct
	gateOpen    bool
	onHoldSince uint32
}

/*
 * Noise gate audio processing.
 */
func (this *noiseGate) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	levelOpen, _ := this.getNumericValue("threshold_open")
	levelClose, _ := this.getNumericValue("threshold_close")
	holdTime, _ := this.getNumericValue("hold_time")
	this.mutex.RUnlock()
	facOpen := decibelsToFactor(levelOpen)
	facClose := decibelsToFactor(levelClose)

	/*
	 * If opening threshold lies BELOW closing threshold, bypass the gate altogether,
	 * but still keep it open.
	 */
	if levelOpen < levelClose {
		copy(out, in)
		this.gateOpen = true
		this.onHoldSince = 0
	} else {
		holdTimeFloat := float64(holdTime)
		holdTimeSeconds := 0.001 * holdTimeFloat
		sampleRateFloat := float64(sampleRate)
		holdSamplesFloat := math.Floor((holdTimeSeconds * sampleRateFloat) + 0.5)
		holdSamples := uint32(holdSamplesFloat)
		gateOpen := this.gateOpen
		onHoldSince := this.onHoldSince

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			amplitude := math.Abs(sample)

			/*
			 * Check if amplitude is above opening threshold.
			 */
			if amplitude > facOpen {
				gateOpen = true
			}

			/*
			 * Check if amplitude is above closing threshold.
			 */
			if amplitude > facClose {
				onHoldSince = 0
			}

			/*
			 * If we're on hold for too long, close the gate.
			 */
			if onHoldSince >= holdSamples {
				gateOpen = false
			}

			fac := float64(0.0)

			/*
			 * Check if gate is open.
			 */
			if gateOpen {
				fac = 1.0
			}

			out[i] = fac * sample

			/*
			 * Increment time on hold, unless it overflows.
			 */
			if onHoldSince < math.MaxUint32 {
				onHoldSince++
			}

		}

		this.gateOpen = gateOpen
		this.onHoldSince = onHoldSince
	}

}

/*
 * Create a noise gate effects unit.
 */
func createNoiseGate() Unit {

	/*
	 * Create effects unit.
	 */
	u := noiseGate{
		unitStruct: unitStruct{
			unitType: UNIT_NOISEGATE,
			params: []Parameter{
				Parameter{
					Name:               "threshold_open",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -20,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "threshold_close",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       -40,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "hold_time",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            1000,
					NumericValue:       50,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
			},
		},
	}

	return &u
}
