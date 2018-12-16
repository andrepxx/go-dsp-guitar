package tuner

import (
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/circular"
	"github.com/andrepxx/go-dsp-guitar/fft"
	"math"
	"math/cmplx"
	"sync"
)

/*
 * Global constants.
 */
const (
	NUM_SAMPLES = 96000
)

/*
 * Data structure representing a musical note.
 */
type noteStruct struct {
	name      string
	frequency float64
}

/*
 * Data structure representing the result of a spectral analysis.
 */
type resultStruct struct {
	cents     int8
	frequency float64
	note      string
}

/*
 * The result of a spectral analysis.
 */
type Result interface {
	Cents() int8
	Frequency() float64
	Note() string
}

/*
 * Data structure representing a tuner.
 */
type tunerStruct struct {
	notes          []noteStruct
	mutexBuffer    sync.RWMutex
	buffer         circular.Buffer
	sampleRate     uint32
	mutexAnalyze   sync.Mutex
	bufCorrelation []float64
	bufFFT         []complex128
}

/*
 * A chromatic instrument tuner.
 */
type Tuner interface {
	Analyze() (Result, error)
	Process(samples []float64, sampleRate uint32)
}

/*
 * Generates a list of notes and their frequencies.
 *
 * f(n) = 2^(n / 12) * 440
 *
 * Where n is the number of half-tone steps relative to A4.
 */
func generateNotes() []noteStruct {

	/*
	 * Create a list of appropriate notes.
	 */
	notes := []noteStruct{
		noteStruct{
			name:      "H1",
			frequency: 61.7354,
		},
		noteStruct{
			name:      "C2",
			frequency: 65.4064,
		},
		noteStruct{
			name:      "C#2",
			frequency: 69.2957,
		},
		noteStruct{
			name:      "D2",
			frequency: 73.4162,
		},
		noteStruct{
			name:      "D#2",
			frequency: 77.7817,
		},
		noteStruct{
			name:      "E2",
			frequency: 82.4069,
		},
		noteStruct{
			name:      "F2",
			frequency: 87.3071,
		},
		noteStruct{
			name:      "F#2",
			frequency: 92.4986,
		},
		noteStruct{
			name:      "G2",
			frequency: 97.9989,
		},
		noteStruct{
			name:      "G#2",
			frequency: 103.8262,
		},
		noteStruct{
			name:      "A2",
			frequency: 110.0000,
		},
		noteStruct{
			name:      "A#2",
			frequency: 116.5409,
		},
		noteStruct{
			name:      "H2",
			frequency: 123.4708,
		},
		noteStruct{
			name:      "C3",
			frequency: 130.8128,
		},
		noteStruct{
			name:      "C#3",
			frequency: 138.5913,
		},
		noteStruct{
			name:      "D3",
			frequency: 146.8324,
		},
		noteStruct{
			name:      "D#3",
			frequency: 155.5635,
		},
		noteStruct{
			name:      "E3",
			frequency: 164.8138,
		},
		noteStruct{
			name:      "F3",
			frequency: 174.6141,
		},
		noteStruct{
			name:      "F#3",
			frequency: 184.9972,
		},
		noteStruct{
			name:      "G3",
			frequency: 195.9978,
		},
		noteStruct{
			name:      "G#3",
			frequency: 207.6523,
		},
		noteStruct{
			name:      "A3",
			frequency: 220.0000,
		},
		noteStruct{
			name:      "A#3",
			frequency: 233.0819,
		},
		noteStruct{
			name:      "H3",
			frequency: 246.9417,
		},
		noteStruct{
			name:      "C4",
			frequency: 261.6256,
		},
		noteStruct{
			name:      "C#4",
			frequency: 277.1826,
		},
		noteStruct{
			name:      "D4",
			frequency: 293.6648,
		},
		noteStruct{
			name:      "D#4",
			frequency: 311.1270,
		},
		noteStruct{
			name:      "E4",
			frequency: 329.6276,
		},
		noteStruct{
			name:      "F4",
			frequency: 349.2282,
		},
		noteStruct{
			name:      "F#4",
			frequency: 369.9944,
		},
		noteStruct{
			name:      "G4",
			frequency: 391.9954,
		},
		noteStruct{
			name:      "G#4",
			frequency: 415.3047,
		},
		noteStruct{
			name:      "A4",
			frequency: 440.0000,
		},
		noteStruct{
			name:      "A#4",
			frequency: 466.1638,
		},
		noteStruct{
			name:      "H4",
			frequency: 493.8833,
		},
		noteStruct{
			name:      "C5",
			frequency: 523.2511,
		},
		noteStruct{
			name:      "C#5",
			frequency: 554.3653,
		},
		noteStruct{
			name:      "D5",
			frequency: 587.3295,
		},
		noteStruct{
			name:      "D#5",
			frequency: 622.2540,
		},
		noteStruct{
			name:      "E5",
			frequency: 659.2551,
		},
		noteStruct{
			name:      "F5",
			frequency: 698.4565,
		},
		noteStruct{
			name:      "F#5",
			frequency: 739.9888,
		},
		noteStruct{
			name:      "G5",
			frequency: 783.9909,
		},
		noteStruct{
			name:      "G#5",
			frequency: 830.6094,
		},
		noteStruct{
			name:      "A5",
			frequency: 880.0000,
		},
		noteStruct{
			name:      "A#5",
			frequency: 932.3275,
		},
		noteStruct{
			name:      "H5",
			frequency: 987.7666,
		},
		noteStruct{
			name:      "C6",
			frequency: 1046.5023,
		},
		noteStruct{
			name:      "C#6",
			frequency: 1108.7305,
		},
		noteStruct{
			name:      "D6",
			frequency: 1174.6591,
		},
		noteStruct{
			name:      "D#6",
			frequency: 1244.5079,
		},
		noteStruct{
			name:      "E6",
			frequency: 1318.5102,
		},
		noteStruct{
			name:      "F6",
			frequency: 1396.9129,
		},
		noteStruct{
			name:      "F#6",
			frequency: 1479.9777,
		},
		noteStruct{
			name:      "G6",
			frequency: 1567.9817,
		},
		noteStruct{
			name:      "G#6",
			frequency: 1661.2188,
		},
		noteStruct{
			name:      "A6",
			frequency: 1760.0000,
		},
		noteStruct{
			name:      "A#6",
			frequency: 1864.6550,
		},
		noteStruct{
			name:      "H6",
			frequency: 1975.5332,
		},
	}

	return notes
}

/*
 * Find the maximum value in a buffer.
 */
func findMaximum(buf []float64) (float64, int) {
	maxVal := math.Inf(-1)
	maxIdx := int(-1)

	/*
	 * Iterate over the buffer and find the maximum value.
	 */
	for idx, value := range buf {

		/*
		 * If we found a value which is greater than any value we
		 * encountered so far, make it the new candidate.
		 */
		if value > maxVal {
			maxVal = value
			maxIdx = idx
		}

	}

	return maxVal, maxIdx
}

/*
 * Returns the deviation from the reference note in cents.
 */
func (this *resultStruct) Cents() int8 {
	return this.cents
}

/*
 * Returns the fundamental frequency of the signal.
 */
func (this *resultStruct) Frequency() float64 {
	return this.frequency
}

/*
 * Returns the name of the closest note on the chromatic scale.
 */
func (this *resultStruct) Note() string {
	return this.note
}

/*
 * Analyze buffered stream for spectral content.
 */
func (this *tunerStruct) Analyze() (Result, error) {
	this.mutexAnalyze.Lock()
	circularBuffer := this.buffer
	bufCorrelation := this.bufCorrelation
	bufCorrrlationLength := len(bufCorrelation)
	bufCorrelationLength64 := uint64(bufCorrrlationLength)
	bufFFT := this.bufFFT
	bufFFTLength := len(bufFFT)
	bufFFTLength64 := uint64(bufFFTLength)
	n := circularBuffer.Length()
	twoN := uint64(2 * n)
	fftSize, _ := fft.NextPowerOfTwo(twoN)

	/*
	 * Ensure that correlation buffer is of correct length.
	 */
	if bufCorrelationLength64 != fftSize {
		bufCorrelation = make([]float64, fftSize)
		this.bufCorrelation = bufCorrelation
	}

	/*
	 * Ensure that FFT buffer is of correct length.
	 */
	if bufFFTLength64 != fftSize {
		bufFFT = make([]complex128, fftSize)
		this.bufFFT = bufFFT
	}

	signalBuffer := bufCorrelation[0:n]
	this.mutexBuffer.RLock()
	sampleRate := this.sampleRate
	err := circularBuffer.Retrieve(signalBuffer)
	this.mutexBuffer.RUnlock()

	/*
	 * Verify that buffer contents could be retrieved.
	 */
	if err != nil {
		msg := err.Error()
		this.mutexAnalyze.Unlock()
		return nil, fmt.Errorf("Failed to retrieve contents of circular buffer: %s", msg)
	} else {
		tailBuffer := bufCorrelation[n:fftSize]
		fft.ZeroFloat(tailBuffer)
		err = fft.RealFourier(bufCorrelation, bufFFT, fft.SCALING_DEFAULT)

		/*
		 * Verify that the forward FFT was calculated successfully.
		 */
		if err != nil {
			msg := err.Error()
			this.mutexAnalyze.Unlock()
			return nil, fmt.Errorf("Failed to calculate forward FFT: %s", msg)
		} else {

			/*
			 * Multiply each element of the spectrum with its complex conjugate.
			 */
			for i, elem := range bufFFT {
				elemConj := cmplx.Conj(elem)
				bufFFT[i] = elem * elemConj
			}

			err = fft.RealInverseFourier(bufFFT, bufCorrelation, fft.SCALING_DEFAULT)

			/*
			 * Verify that the inverse FFT was calculated successfully.
			 */
			if err != nil {
				msg := err.Error()
				this.mutexAnalyze.Unlock()
				return nil, fmt.Errorf("Failed to calculate inverse FFT: %s", msg)
			} else {
				notes := this.notes
				noteCount := len(notes)
				lastNote := noteCount - 1
				lowFreq := notes[0].frequency
				highFreq := notes[lastNote].frequency
				sampleRateFloat := float64(sampleRate)
				lowIdx := int((sampleRateFloat / highFreq) + 0.5)
				lowIdx64 := uint64(lowIdx)

				/*
				 * This might happen when the float value is infinite.
				 */
				if (lowIdx < 0) || (lowIdx64 >= twoN) {
					lowIdx = 0
					lowIdx64 = 0
				}

				highIdx := int((sampleRateFloat / lowFreq) + 0.5)
				highIdx64 := uint64(highIdx)

				/*
				 * This might happen when the float value is infinite.
				 */
				if (highIdx < 0) || (highIdx64 >= twoN) {
					maxIdx := twoN - 1
					highIdx = int(maxIdx)
					highIdx64 = maxIdx
				}

				subCorrelation := bufCorrelation[lowIdx:highIdx]
				maxVal, maxIdx := findMaximum(subCorrelation)
				idx := lowIdx + maxIdx
				idxUp := idx + 1

				/*
				 * Prevent overrun.
				 */
				if idxUp > n {
					idxUp = n
				}

				idxDown := idx - 1

				/*
				 * Prevent underrun.
				 */
				if idxDown < 0 {
					idxDown = 0
				}

				valueLeft := bufCorrelation[idxDown]
				valueRight := bufCorrelation[idxUp]
				idxFloat := float64(idx)
				valueDiff := valueRight - valueLeft
				valueSum := valueRight + valueLeft
				halfDiff := 0.5 * valueDiff
				doubleMaxVal := 2.0 * maxVal
				denominatorDiff := doubleMaxVal - valueSum
				shiftEstimation := halfDiff / denominatorDiff

				/*
				 * Limit shift estimation to plus/minus half a sample.
				 */
				if shiftEstimation < -0.5 {
					shiftEstimation = -0.5
				} else if shiftEstimation > 0.5 {
					shiftEstimation = 0.5
				}

				idxFloat += shiftEstimation
				actualFrequency := sampleRateFloat / idxFloat
				actualNote := "Unknown"
				actualCents := math.Inf(1)
				actualCentsAbs := math.Abs(actualCents)

				/*
				 * Iterate over all notes and find the closest match.
				 */
				for _, note := range notes {
					freq := note.frequency
					freqRatio := actualFrequency / freq
					diffCents := 1200.0 * math.Log2(freqRatio)
					diffCentsAbs := math.Abs(diffCents)

					/*
					 * If this is the closest we've seen so far, make this the best match.
					 */
					if diffCentsAbs < actualCentsAbs {
						actualNote = note.name
						actualCents = diffCents
						actualCentsAbs = diffCentsAbs
					}

				}

				actualCentsInfinite := math.IsInf(actualCents, 0)
				actualCentsNaN := math.IsNaN(actualCents)
				actualCentsInt := int8(0)

				/*
				 * If cents are finite, use them.
				 */
				if !(actualCentsInfinite || actualCentsNaN) {
					actualCentsInt = int8(actualCents)
				}

				/*
				 * Create result of signal analysis.
				 */
				result := resultStruct{
					cents:     actualCentsInt,
					frequency: actualFrequency,
					note:      actualNote,
				}

				this.mutexAnalyze.Unlock()
				return &result, nil
			}

		}

	}

}

/*
 * Stream samples for later analysis.
 */
func (this *tunerStruct) Process(samples []float64, sampleRate uint32) {
	this.mutexBuffer.Lock()
	this.buffer.Enqueue(samples...)
	this.sampleRate = sampleRate
	this.mutexBuffer.Unlock()
}

/*
 * Creates an instrument tuner.
 */
func Create() Tuner {
	notes := generateNotes()
	buffer := circular.CreateBuffer(NUM_SAMPLES)

	/*
	 * Create data structure for a guitar tuner.
	 */
	t := tunerStruct{
		notes:  notes,
		buffer: buffer,
	}

	return &t
}
