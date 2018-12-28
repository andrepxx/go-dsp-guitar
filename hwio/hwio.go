package hwio

import (
	"fmt"
	"github.com/andrepxx/go-jack"
	"strconv"
	"sync"
)

/*
 * Function pointer for implementing signal processors.
 */
type Processor func([][]float64, [][]float64, uint32)

/*
 * Function pointer for implementing sample rate listeners.
 */
type SampleRateListener func(uint32)

/*
 * Data structure representing a handle to a hardware input and output and an
 * associated signal processor.
 */
type Binding struct {
	inputs    []*jack.Port
	outputs   []*jack.Port
	processor Processor
	listener  SampleRateListener
}

/*
 * Global constants.
 */
const (
	INPUT_CHANNELS  = 2
	OUTPUT_CHANNELS = INPUT_CHANNELS + 3
)

/*
 * Global variables.
 */
var g_client *jack.Client       // JACK client handle.
var g_mutex sync.RWMutex        // Mutex for bindings.
var g_bindings []*Binding = nil // All currently active bindings.
var g_inputBuffers [][]float64  // Input buffers.
var g_outputBuffers [][]float64 // Output buffers.
var g_sampleRate uint32         // Sample rate.

/*
 * Convert audio samples to floating-point numbers.
 */
func samplesToFloats(in []jack.AudioSample, out []float64) error {

	/*
	 * Verify that the output buffer has an appropriate size
	 */
	if len(out) < len(in) {
		return fmt.Errorf("%s", "Cannot convert samples to floats: Output buffer is too small.")
	} else {

		/*
		 * Convert each audio sample to a floating-point number.
		 */
		for i, sample := range in {
			out[i] = float64(sample)
		}

		return nil
	}

}

/*
 * Convert floating-point numbers to audio samples.
 */
func floatsToSamples(in []float64, out []jack.AudioSample) error {

	/*
	 * Verify that the output buffer has an appropriate size
	 */
	if len(out) < len(in) {
		return fmt.Errorf("%s", "Cannot convert floats to samples: Output buffer is too small.")
	} else {

		/*
		 * Convert each floating-point number to an audio sample.
		 */
		for i, sample := range in {
			out[i] = jack.AudioSample(sample)
		}

		return nil
	}

}

/*
 * Interrupt handler called when the hardware has audio to process.
 */
func process(nframes uint32) int {
	g_mutex.RLock()

	/*
	 * Process audio for each binding.
	 */
	for _, binding := range g_bindings {
		inputs := binding.inputs
		outputs := binding.outputs

		/*
		 * Read audio from each input channel.
		 */
		for i, input := range inputs {
			hwInputBuffer := input.GetBuffer(nframes)
			bufferSize := len(hwInputBuffer)

			/*
			 * Ensure the size of the current input buffer matches the size of the hardware buffer.
			 */
			if len(g_inputBuffers[i]) != bufferSize {
				g_inputBuffers[i] = make([]float64, bufferSize)
			}

			err := samplesToFloats(hwInputBuffer, g_inputBuffers[i])

			/*
			 * If conversion failed, log error, otherwise perform processing.
			 */
			if err != nil {
				msg := err.Error()
				fmt.Printf("Error in real-time thread: %s", msg)
			}

		}

		/*
		 * Prepare output buffer for each output channel.
		 */
		for i, output := range outputs {
			hwOutputBuffer := output.GetBuffer(nframes)
			bufferSize := len(hwOutputBuffer)

			/*
			 * Ensure the size of the current output buffer matches the size of the hardware buffer.
			 */
			if len(g_outputBuffers[i]) != bufferSize {
				g_outputBuffers[i] = make([]float64, bufferSize)
			}

		}

		binding.processor(g_inputBuffers, g_outputBuffers, g_sampleRate)

		/*
		 * Write audio to each output channel.
		 */
		for i, output := range outputs {
			hwOutputBuffer := output.GetBuffer(nframes)
			err := floatsToSamples(g_outputBuffers[i], hwOutputBuffer)

			/*
			 * If conversion failed, log error.
			 */
			if err != nil {
				msg := err.Error()
				fmt.Printf("Error in real-time thread: %s", msg)
			}

		}

	}

	g_mutex.RUnlock()
	return 0
}

/*
 * Interrupt handler called when the hardware adjusts the sample rate.
 */
func sampleRate(rate uint32) int {
	g_sampleRate = rate

	/*
	 * Notify each binding about the change.
	 */
	for _, binding := range g_bindings {
		binding.listener(rate)
	}

	return 0
}

/*
 * Initialize the hardware for signal processing.
 */
func initialize() (*jack.Client, error) {
	client, _ := jack.ClientOpen("go-dsp-guitar", jack.NoStartServer)

	/*
	 * Check if we are connected to the JACK server.
	 */
	if client == nil {
		return nil, fmt.Errorf("%s", "Could not connect to JACK server.")
	} else {
		statusProcess := client.SetProcessCallback(process)

		/*
		 * Check if we could register our application as a signal processor.
		 */
		if statusProcess != 0 {
			return nil, fmt.Errorf("%s", "Failed to set process callback.")
		} else {
			statusSampleRate := client.SetSampleRateCallback(sampleRate)

			/*
			 * Check if we could register a sample rate callback.
			 */
			if statusSampleRate != 0 {
				return nil, fmt.Errorf("%s", "Failed to set sample rate callback.")
			} else {
				statusActivate := client.Activate()

				/*
				 * Check if we could activate JACK.
				 */
				if statusActivate != 0 {
					return nil, fmt.Errorf("%s", "Failed to activate client.")
				} else {
					return client, nil
				}

			}

		}

	}

}

/*
 * Get DSP load.
 */
func DSPLoad() float32 {
	res := float32(0.0)
	g_mutex.RLock()

	/*
	 * Check if client is registered.
	 */
	if g_client != nil {
		res = g_client.CPULoad()
	}

	g_mutex.RUnlock()
	return res
}

/*
 * Get frames per period.
 */
func FramesPerPeriod() uint32 {
	res := uint32(0)
	g_mutex.RLock()

	/*
	 * Check if client is registered.
	 */
	if g_client != nil {
		res = g_client.GetBufferSize()
	}

	g_mutex.RUnlock()
	return res
}

/*
 * Register a binding to a hardware interface.
 */
func Register(processor Processor, listener SampleRateListener) (*Binding, error) {
	err := error(nil)
	g_mutex.RLock()

	/*
	 * If no bindings exist yet, initialize hardware first.
	 */
	if g_bindings == nil {
		g_mutex.RUnlock()
		g_mutex.Lock()
		g_client, err = initialize()
		g_bindings = []*Binding{}
		g_inputBuffers = make([][]float64, INPUT_CHANNELS)
		g_outputBuffers = make([][]float64, OUTPUT_CHANNELS)
		g_mutex.Unlock()
		g_mutex.RLock()
	}

	g_mutex.RUnlock()

	/*
	 * Check, whether hardware was initialized successfully.
	 */
	if err != nil {
		return nil, err
	} else {
		inputs := make([]*jack.Port, INPUT_CHANNELS)
		outputs := make([]*jack.Port, OUTPUT_CHANNELS)

		/*
		 * Create input and output for each input channel.
		 */
		for idx, _ := range inputs {
			idxLong := int64(idx)
			sChannelNumber := strconv.FormatInt(idxLong, 10)
			inputName := "in_" + sChannelNumber
			inputs[idx] = g_client.PortRegister(inputName, jack.DEFAULT_AUDIO_TYPE, jack.PortIsInput, 0)
			outputName := "out_" + sChannelNumber
			outputs[idx] = g_client.PortRegister(outputName, jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)
		}

		/*
		 * Names of additional channels to register.
		 */
		additionalChannels := []string{
			"master_left",
			"master_right",
			"metronome",
		}

		nAdditional := len(additionalChannels)
		baseIdx := OUTPUT_CHANNELS - nAdditional

		/*
		 * Register additional channels.
		 */
		for i, additionalChannel := range additionalChannels {
			idx := baseIdx + i
			outputs[idx] = g_client.PortRegister(additionalChannel, jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)
		}

		/*
		 * Create hardware binding.
		 */
		binding := &Binding{
			inputs:    inputs,
			outputs:   outputs,
			processor: processor,
			listener:  listener,
		}

		g_mutex.Lock()
		g_bindings = append(g_bindings, binding)
		g_mutex.Unlock()
		sampleRate(g_sampleRate)
		return binding, nil
	}

}

/*
 * Set frames per period.
 */
func SetFramesPerPeriod(n uint32) {
	g_mutex.RLock()

	/*
	 * Check if client is registered.
	 */
	if g_client != nil {
		g_client.SetBufferSize(n)
	}

	g_mutex.RUnlock()
}

/*
 * Unregister a binding to a hardware interface.
 */
func Unregister(binding *Binding) {
	idx := int(-1)
	g_mutex.RLock()

	/*
	 * Iterate over the bindings.
	 */
	for i, b := range g_bindings {

		/*
		 * Check if we have the binding we're about to remove.
		 */
		if b == binding {
			idx = i
		}

	}

	/*
	 * If we found the binding, remove it.
	 */
	if idx > 0 {
		inputs := binding.inputs
		outputs := binding.outputs
		idxInc := idx + 1
		g_mutex.RUnlock()
		g_mutex.Lock()

		/*
		 * Unregister all input ports.
		 */
		for _, port := range inputs {
			g_client.PortUnregister(port)
		}

		/*
		 * Unregister all output ports.
		 */
		for _, port := range outputs {
			g_client.PortUnregister(port)
		}

		g_bindings = append(g_bindings[:idx], g_bindings[idxInc:]...)
		g_mutex.Unlock()
		g_mutex.RLock()
	}

	/*
	 * If no bindings exist, terminate connection to JACK.
	 */
	if len(g_bindings) == 0 {
		g_mutex.RUnlock()
		g_mutex.Lock()
		g_client.Close()
		g_client = nil
		g_bindings = nil
		g_mutex.Unlock()
		g_mutex.RLock()
	}

	g_mutex.RUnlock()
}

/*
 * Connects a source port to a destination port.
 */
func Connect(sourcePort string, destinationPort string) {
	g_mutex.RLock()

	/*
	 * Check if client is registered.
	 */
	if g_client != nil {
		g_client.Connect(sourcePort, destinationPort)
	}

	g_mutex.RUnlock()
}
