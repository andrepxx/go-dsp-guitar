package effects

import (
	"github.com/andrepxx/go-dsp-guitar/random"
	"math"
)

/*
 * Data structure representing a signal generator.
 */
type signalGenerator struct {
	unitStruct
	phase float64
	prng  random.PseudoRandomNumberGenerator
}

/*
 * Signal generator audio processing.
 */
func (this *signalGenerator) Process(in []float64, out []float64, sampleRate uint32) {
	this.mutex.RLock()
	inputAmplitude, _ := this.getNumericValue("input_amplitude")
	inputGain, _ := this.getNumericValue("input_gain")
	signalType, _ := this.getDiscreteValue("signal_type")
	signalFrequency, _ := this.getNumericValue("signal_frequency")
	signalAmplitude, _ := this.getNumericValue("signal_amplitude")
	signalGain, _ := this.getNumericValue("signal_gain")
	this.mutex.RUnlock()
	inputAmplitudeFloat := float64(inputAmplitude)
	facInputGain := decibelsToFactor(inputGain)
	facInput := (0.01 * inputAmplitudeFloat) * facInputGain
	facSignalGain := decibelsToFactor(signalGain)
	signalAmplitudeFloat := float64(signalAmplitude)
	facSignal := (0.01 * signalAmplitudeFloat) * facSignalGain
	phase := this.phase
	signalFrequencyFloat := float64(signalFrequency)
	sampleRateFloat := float64(sampleRate)
	phaseIncrement := MATH_TWO_PI * (signalFrequencyFloat / sampleRateFloat)
	twoOverPi := 2.0 / math.Pi
	n := len(in)
	nFloat := float64(n)

	/*
	 * Generate the appropriate signal.
	 */
	switch signalType {
	case "sine":

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			iFloat := float64(i)
			updatedPhase := phase + (iFloat * phaseIncrement)
			currentPhase := math.Mod(updatedPhase, MATH_TWO_PI)
			signal := math.Sin(currentPhase)
			out[i] = (facInput * sample) + (facSignal * signal)
		}

		phase += nFloat * phaseIncrement
		phase = math.Mod(phase, MATH_TWO_PI)
		break
	case "triangle":

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			iFloat := float64(i)
			updatedPhase := phase + (iFloat * phaseIncrement)
			currentPhase := math.Mod(updatedPhase, MATH_TWO_PI)
			signal := 0.0

			/*
			 * Check whether the waveform is rising or falling.
			 */
			if currentPhase < math.Pi {
				signal = (twoOverPi * currentPhase) - 1.0
			} else {
				signal = 3.0 - (twoOverPi * currentPhase)
			}

			out[i] = (facInput * sample) + (facSignal * signal)
		}

		phase += nFloat * phaseIncrement
		phase = math.Mod(phase, MATH_TWO_PI)
		break
	case "square":

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			iFloat := float64(i)
			updatedPhase := phase + (iFloat * phaseIncrement)
			currentPhase := math.Mod(updatedPhase, MATH_TWO_PI)
			signal := signFloat(math.Pi - currentPhase)
			out[i] = (facInput * sample) + (facSignal * signal)
		}

		phase += nFloat * phaseIncrement
		phase = math.Mod(phase, MATH_TWO_PI)
		break
	case "sawtooth":

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			iFloat := float64(i)
			updatedPhase := phase + (iFloat * phaseIncrement)
			currentPhase := math.Mod(updatedPhase, MATH_TWO_PI)
			signal := currentPhase / math.Pi

			/*
			 * Check whether we're after the phase jump.
			 */
			if currentPhase > math.Pi {
				signal -= 2.0
			}

			out[i] = (facInput * sample) + (facSignal * signal)
		}

		phase += nFloat * phaseIncrement
		phase = math.Mod(phase, MATH_TWO_PI)
		break
	case "noise":
		prng := this.prng

		/*
		 * Check if pseudo-random number generator is initialized.
		 */
		if prng == nil {
			prng = random.CreatePRNG(1337)
			this.prng = prng
		}

		/*
		 * Process each sample.
		 */
		for i, sample := range in {
			r := prng.NextFloat()
			uniform := (1.0 - (2.0 * r))
			out[i] = (facInput * sample) + (facSignal * uniform)
		}

		break
	}

	this.phase = phase
}

/*
 * Create a signal generator effects unit.
 */
func createSignalGenerator() Unit {

	/*
	 * Create effects unit.
	 */
	u := signalGenerator{
		unitStruct: unitStruct{
			unitType: UNIT_SIGNALGENERATOR,
			params: []Parameter{
				Parameter{
					Name:               "input_amplitude",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "input_gain",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
					Maximum:            0,
					NumericValue:       0,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "signal_type",
					Type:               PARAMETER_TYPE_DISCRETE,
					Minimum:            -1,
					Maximum:            -1,
					NumericValue:       -1,
					DiscreteValueIndex: 0,
					DiscreteValues: []string{
						"sine",
						"triangle",
						"square",
						"sawtooth",
						"noise",
					},
				},
				Parameter{
					Name:               "signal_frequency",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            1,
					Maximum:            20000,
					NumericValue:       440,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "signal_amplitude",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            0,
					Maximum:            100,
					NumericValue:       100,
					DiscreteValueIndex: -1,
					DiscreteValues:     nil,
				},
				Parameter{
					Name:               "signal_gain",
					Type:               PARAMETER_TYPE_NUMERIC,
					Minimum:            -60,
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
