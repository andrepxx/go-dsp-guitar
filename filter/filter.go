package filter

import (
	"encoding/json"
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/fft"
	"github.com/andrepxx/go-dsp-guitar/resample"
	"github.com/andrepxx/go-dsp-guitar/wave"
	"io/ioutil"
	"math"
	"math/cmplx"
	"strconv"
)

/*
 * Global constants.
 */
const (
	CHANNEL_COUNT = 1
)

/*
 * Global variables.
 */
var g_sampleRates = []uint32{
	22050,
	32000,
	44100,
	48000,
	88200,
	96000,
	192000,
}

/*
 * Data structure describing an FIR filter.
 */
type filterDescriptorStruct struct {
	Name         string
	Path         string
	Compensation int32
}

/*
 * Data structure containing the coefficients for an FIR filter.
 */
type impulseResponseStruct struct {
	name             string
	sampleRate       uint32
	gainCompensation float64
	data             []float64
}

/*
 * A collection of impulse responses.
 */
type impulseResponsesStruct struct {
	responses []impulseResponseStruct
}

/*
 * Interface type representing a collection of impulse responses.
 */
type ImpulseResponses interface {
	CreateFilter(name string, sampleRate uint32) Filter
	Names() []string
}

/*
 * Data structure implementing an FIR filter.
 */
type filterStruct struct {
	impulseResponse     impulseResponseStruct
	fourierTransform    fft.FourierTransform
	filterComplex       []complex128
	filteredComplex     []complex128
	inputBuffer         []float64
	inputBufferComplex  []complex128
	outputBuffer        []float64
	outputBufferComplex []complex128
	tailBuffer          []float64
}

/*
 * Interface type representing an FIR filter.
 */
type Filter interface {
	Add(other Filter) (Filter, error)
	Coefficients() []float64
	Multiply(scalar float64) Filter
	Normalize() Filter
	Process(inputBuffer []float64, outputBuffer []float64) error
	Reduce(order uint32) Filter
	SampleRate() uint32
}

/*
 * Calculate the complex hadamard product of two vectors.
 */
func hadamardComplex(result []complex128, a []complex128, b []complex128) error {
	L := len(result)
	N := len(a)
	M := len(b)

	/*
	 * Check if buffers are the same size.
	 */
	if (N != M) || (L != N) {
		return fmt.Errorf("%s", "Failed to calculate hadamard product: All buffers must be the same size.")
	} else {

		/*
		 * Multiply the contents of the buffer
		 */
		for i, _ := range result {
			result[i] = a[i] * b[i]
		}

		return nil
	}

}

/*
 * Estimate the gain of an FIR filter.
 */
func estimateGain(coefficients []float64) float64 {
	sum := 0.0

	/*
	 * Sum the squares of the filter coefficients.
	 */
	for _, coefficient := range coefficients {
		sum += coefficient * coefficient
	}

	return math.Sqrt(sum)
}

/*
 * Find the maximum absolute value in a vector.
 */
func peakValue(vec []float64) float64 {
	peak := 0.0

	/*
	 * Iterate over all values in the vector.
	 */
	for _, value := range vec {
		abs := math.Abs(value)

		/*
		 * If we found a larger absolute value, keep it.
		 */
		if abs > peak {
			peak = abs
		}

	}

	return peak
}

/*
 * Add another filter to this one.
 */
func (this *filterStruct) Add(other Filter) (Filter, error) {

	/*
	 * Check if the filter to be added is nil.
	 */
	if other == nil {
		return this, nil
	} else {
		otherStruct := other.(*filterStruct)
		irA := this.impulseResponse
		irB := otherStruct.impulseResponse
		rateA := irA.sampleRate
		rateB := irB.sampleRate

		/*
		 * Only filters with the same sample rate can be added.
		 */
		if rateA != rateB {
			return nil, fmt.Errorf("%s", "Cannot add filters: Sample rates do not match.")
		} else {
			coeffsA := irA.data
			coeffsB := irB.data
			nA := len(coeffsA)
			nB := len(coeffsB)
			nResult := nA

			/*
			 * Check if the other filter is larger.
			 */
			if nB > nResult {
				nResult = nB
			}

			coeffsResult := make([]float64, nResult)
			copy(coeffsResult, coeffsA)

			/*
			 * Add filter coefficients of other filter.
			 */
			for i, coeff := range coeffsB {
				coeffsResult[i] += coeff
			}

			nameA := irA.name
			nameB := irB.name
			nameResult := nameA + " + " + nameB

			/*
			 * Create the resulting impulse response.
			 */
			irResult := impulseResponseStruct{
				name:             nameResult,
				sampleRate:       rateA,
				gainCompensation: 0.0,
				data:             coeffsResult,
			}

			ft := fft.CreateFourierTransform()
			bufFilterC := make([]complex128, 0)
			bufFilteredC := make([]complex128, 0)
			bufInput := make([]float64, 0)
			bufInputC := make([]complex128, 0)
			bufOutput := make([]float64, 0)
			bufOutputC := make([]complex128, 0)
			bufTail := make([]float64, 0)

			/*
			 * Create a new filter.
			 */
			fltFilter := filterStruct{
				impulseResponse:     irResult,
				fourierTransform:    ft,
				filterComplex:       bufFilterC,
				filteredComplex:     bufFilteredC,
				inputBuffer:         bufInput,
				inputBufferComplex:  bufInputC,
				outputBuffer:        bufOutput,
				outputBufferComplex: bufOutputC,
				tailBuffer:          bufTail,
			}

			return &fltFilter, nil
		}

	}

}

/*
 * Return filter coefficients.
 */
func (this *filterStruct) Coefficients() []float64 {
	ir := this.impulseResponse
	coeff := ir.data
	size := len(coeff)
	coeffCopy := make([]float64, size)
	copy(coeffCopy, coeff)
	return coeffCopy
}

/*
 * Multiply the filter with a scalar factor.
 */
func (this *filterStruct) Multiply(scalar float64) Filter {
	scalarString := strconv.FormatFloat(scalar, 'f', -1, 64)
	ir := this.impulseResponse
	coeffs := ir.data
	n := len(coeffs)
	coeffsResult := make([]float64, n)

	/*
	 * Calculate the new filter coefficients with compensated gain.
	 */
	for i, coeff := range coeffs {
		coeffsResult[i] = scalar * coeff
	}

	sr := ir.sampleRate
	irName := ir.name
	irNewName := scalarString + " * (" + irName + ")"

	/*
	 * Create a new impulse response structure.
	 */
	irResult := impulseResponseStruct{
		name:             irNewName,
		sampleRate:       sr,
		gainCompensation: 0.0,
		data:             coeffsResult,
	}

	ft := fft.CreateFourierTransform()
	bufFilterC := make([]complex128, 0)
	bufFilteredC := make([]complex128, 0)
	bufInput := make([]float64, 0)
	bufInputC := make([]complex128, 0)
	bufOutput := make([]float64, 0)
	bufOutputC := make([]complex128, 0)
	bufTail := make([]float64, 0)

	/*
	 * Create a new filter.
	 */
	fltFilter := filterStruct{
		impulseResponse:     irResult,
		fourierTransform:    ft,
		filterComplex:       bufFilterC,
		filteredComplex:     bufFilteredC,
		inputBuffer:         bufInput,
		inputBufferComplex:  bufInputC,
		outputBuffer:        bufOutput,
		outputBufferComplex: bufOutputC,
		tailBuffer:          bufTail,
	}

	return &fltFilter
}

/*
 * Normalize the filter to compensate for gain.
 */
func (this *filterStruct) Normalize() Filter {
	ir := this.impulseResponse
	coeffs := ir.data
	gain := estimateGain(coeffs)
	compensation := ir.gainCompensation
	fac := compensation / gain
	fltFilter := this.Multiply(fac)
	return fltFilter
}

/*
 * Reads samples from the input buffer, passes them through the filter and writes
 * samples to the output buffer.
 */
func (this *filterStruct) Process(inputBuffer []float64, outputBuffer []float64) error {
	N := len(inputBuffer)
	M := len(outputBuffer)

	/*
	 * Check if output and input buffer are the same size.
	 */
	if M != N {
		return fmt.Errorf("%s", "Output and input buffer must be of the same size.")
	} else {
		ir := this.impulseResponse
		coefficients := ir.data

		/*
		 * Check if impulse response exists.
		 */
		if coefficients == nil {
			return fmt.Errorf("%s", "Impulse response must not be nil.")
		} else {
			L := len(coefficients)

			/*
			 * Check if filter is empty.
			 */
			if L == 0 {
				fft.ZeroFloat(outputBuffer)
			} else {
				ft := this.fourierTransform
				N64 := uint64(N)
				L64 := uint64(L)
				Npower, _ := fft.NextPowerOfTwo(N64)
				blockSize, _ := fft.NextPowerOfTwo(L64)
				numBlocks := Npower / blockSize
				overflow := Npower % blockSize

				/*
				 * If there is overflow, add another block.
				 */
				if overflow != 0 {
					numBlocks++
				}

				/*
				 * Process each block
				 */
				for i := uint64(0); i < numBlocks; i++ {
					fftSize64 := blockSize << 1
					fftSize := int(fftSize64)
					filterComplex := this.filterComplex

					/*
					 * Pre-calculate the FFT of the filter.
					 */
					if len(filterComplex) != fftSize {
						coefficientsPadded := make([]float64, fftSize)
						copy(coefficientsPadded[0:L], coefficients)
						filterComplex = make([]complex128, fftSize)
						ft.RealFourier(coefficientsPadded, filterComplex, fft.SCALING_DEFAULT)
						this.filterComplex = filterComplex
					}

					filteredComplex := this.filteredComplex

					/*
					 * Check if complex-valued filtered (FFT) buffer is of correct size.
					 */
					if len(filteredComplex) != fftSize {
						filteredComplex = make([]complex128, fftSize)
						this.filteredComplex = filteredComplex
					}

					filterInputBuffer := this.inputBuffer

					/*
					 * Check if real-valued input buffer is of the correct size.
					 */
					if len(filterInputBuffer) != fftSize {
						filterInputBuffer = make([]float64, fftSize)
						this.inputBuffer = filterInputBuffer
					}

					filterOutputBuffer := this.outputBuffer

					/*
					 * Check if real-valued output buffer is of the correct size.
					 */
					if len(filterOutputBuffer) != fftSize {
						filterOutputBuffer = make([]float64, fftSize)
						this.outputBuffer = filterOutputBuffer
					}

					tailBuffer := this.tailBuffer

					/*
					 * Check if real-valued tail buffer is of the correct size.
					 */
					if len(tailBuffer) != fftSize {
						tailBuffer = make([]float64, fftSize)
						this.tailBuffer = tailBuffer
					}

					lBound := i * blockSize
					uBound := lBound + blockSize

					/*
					 * Prevent exceeding upper bound.
					 */
					if uBound > N64 {
						uBound = N64
					}

					currentInputBuffer := inputBuffer[lBound:uBound]
					currentOutputBuffer := outputBuffer[lBound:uBound]
					numSamples := uBound - lBound
					copy(filterInputBuffer[0:numSamples], currentInputBuffer)
					fft.ZeroFloat(filterInputBuffer[numSamples:])
					ft.RealFourier(filterInputBuffer, filteredComplex, fft.SCALING_DEFAULT)
					err := hadamardComplex(filteredComplex, filteredComplex, filterComplex)

					/*
					 * Check if hadamard product was calculated successfully.
					 */
					if err != nil {
						return err
					} else {
						ft.RealInverseFourier(filteredComplex, filterOutputBuffer, fft.SCALING_DEFAULT)

						/*
						 * Calculate the total output by overlapping with the tail of the
						 * previous calculation.
						 */
						for j, elem := range filterOutputBuffer {
							tailElem := tailBuffer[j]
							pre := elem + tailElem
							j64 := uint64(j)

							/*
							 * Write a portion to the current output buffer
							 * and update tail buffer.
							 */
							if j64 < numSamples {

								/*
								 * Ensure that the output is in range.
								 */
								if pre > 1.0 {
									currentOutputBuffer[j] = 1.0
								} else if pre < -1.0 {
									currentOutputBuffer[j] = -1.0
								} else {
									currentOutputBuffer[j] = pre
								}

							} else {
								idx := j64 - numSamples
								tailBuffer[idx] = pre
							}

						}

						tailSize64 := fftSize64 - numSamples
						fft.ZeroFloat(tailBuffer[tailSize64:])
					}

				}

			}

			return nil
		}

	}

}

/*
 * Approximate this filter by one of the given order.
 */
func (this *filterStruct) Reduce(order uint32) Filter {
	ir := this.impulseResponse
	coefficients := ir.data
	orderWord := uint64(order)
	n := len(coefficients)
	nWord := uint64(n)
	nFftSource, _ := fft.NextPowerOfTwo(nWord)
	nFftTarget, _ := fft.NextPowerOfTwo(orderWord)
	coefficientsPadded := make([]float64, nFftSource)
	copy(coefficientsPadded, coefficients)
	nFftSourceWord := uint32(nFftSource)
	nFftTargetWord := uint32(nFftTarget)

	/*
	 * Check if the requested order is exceeded.
	 */
	if nWord <= orderWord {
		return this
	} else {
		ft := this.fourierTransform
		fr := make([]complex128, nFftSource)
		ft.RealFourier(coefficientsPadded, fr, fft.SCALING_DEFAULT)
		numPositiveFreqsSource := (nFftSourceWord >> 1) + 1
		frPos := fr[:numPositiveFreqsSource]
		nFftTargetHalf := nFftTargetWord >> 1
		numPositiveFreqsTarget := nFftTargetHalf + 1
		frPosNew := resample.Frequency(frPos, numPositiveFreqsTarget)
		frNew := make([]complex128, nFftTarget)
		copy(frNew, frPosNew)

		/*
		 * Generate negative frequency values.
		 */
		for i := uint32(1); i < nFftTargetHalf; i++ {
			elem := frPosNew[i]
			elemConj := cmplx.Conj(elem)
			idx := nFftTargetWord - i
			frNew[idx] = elemConj
		}

		targetResponse := make([]float64, nFftTarget)
		ft.RealInverseFourier(frNew, targetResponse, fft.SCALING_DEFAULT)
		coeffsNew := targetResponse[:order]
		nameNew := ir.name + " (" + string(order) + ")"
		rate := ir.sampleRate
		compensation := ir.gainCompensation

		/*
		 * Create a new impulse response structure.
		 */
		irNew := impulseResponseStruct{
			name:             nameNew,
			gainCompensation: compensation,
			sampleRate:       rate,
			data:             coeffsNew,
		}

		ftNewFilter := fft.CreateFourierTransform()
		bufFilterC := make([]complex128, 0)
		bufFilteredC := make([]complex128, 0)
		bufInput := make([]float64, 0)
		bufInputC := make([]complex128, 0)
		bufOutput := make([]float64, 0)
		bufOutputC := make([]complex128, 0)
		bufTail := make([]float64, 0)

		/*
		 * Create a new filter.
		 */
		fltFilter := filterStruct{
			fourierTransform:    ftNewFilter,
			impulseResponse:     irNew,
			filterComplex:       bufFilterC,
			filteredComplex:     bufFilteredC,
			inputBuffer:         bufInput,
			inputBufferComplex:  bufInputC,
			outputBuffer:        bufOutput,
			outputBufferComplex: bufOutputC,
			tailBuffer:          bufTail,
		}

		return &fltFilter
	}

}

/*
 * Returns the sample rate this filter is designed to operate at.
 */
func (this *filterStruct) SampleRate() uint32 {
	ir := this.impulseResponse
	sampleRate := ir.sampleRate
	return sampleRate
}

/*
 * Retrieves an impulse response filter from a collection of impulse responses and
 * creates an FIR filter from it.
 */
func (this *impulseResponsesStruct) CreateFilter(name string, sampleRate uint32) Filter {

	/*
	 * Iterate over the filter collection.
	 */
	for _, ir := range this.responses {

		/*
		 * Check if both name and sample rate match.
		 */
		if (ir.name == name) && (ir.sampleRate == sampleRate) {
			ft := fft.CreateFourierTransform()
			bufFilterC := make([]complex128, 0)
			bufFilteredC := make([]complex128, 0)
			bufInput := make([]float64, 0)
			bufInputC := make([]complex128, 0)
			bufOutput := make([]float64, 0)
			bufOutputC := make([]complex128, 0)
			bufTail := make([]float64, 0)

			/*
			 * Create a new filter.
			 */
			fltFilter := filterStruct{
				impulseResponse:     ir,
				fourierTransform:    ft,
				filterComplex:       bufFilterC,
				filteredComplex:     bufFilteredC,
				inputBuffer:         bufInput,
				inputBufferComplex:  bufInputC,
				outputBuffer:        bufOutput,
				outputBufferComplex: bufOutputC,
				tailBuffer:          bufTail,
			}

			return &fltFilter
		}

	}

	return nil
}

/*
 * Retrieves the names of all impulse responses.
 */
func (this *impulseResponsesStruct) Names() []string {
	names := make([]string, 0)

	/*
	 * Iterate over the filter collection.
	 */
	for _, ir := range this.responses {
		name := ir.name
		contained := false

		/*
		 * Iterate over the names to check whether it's still there.
		 */
		for _, currentName := range names {

			/*
			 * If names match, we already know a version of this impulse response.
			 */
			if currentName == name {
				contained = true
			}

		}

		/*
		 * If this name is not already known, add it to the list.
		 */
		if !contained {
			names = append(names, name)
		}

	}

	return names
}

/*
 * Imports a set of impulse responses using a descriptor file.
 */
func Import(descriptorFilePath string) (ImpulseResponses, error) {
	content, err := ioutil.ReadFile(descriptorFilePath)

	/*
	 * Check if file could be read.
	 */
	if err != nil {
		return nil, fmt.Errorf("Failed to read descriptor file: '%s'", descriptorFilePath)
	} else {
		descriptors := []filterDescriptorStruct{}
		err = json.Unmarshal(content, &descriptors)

		/*
		 * Check if file failed to unmarshal.
		 */
		if err != nil {
			return nil, fmt.Errorf("Failed to decode descriptor file: '%s'", descriptorFilePath)
		} else {
			impulseResponseList := []impulseResponseStruct{}

			/*
			 * Iterate over all filter descriptors and load the corresponding
			 * FIR filter coefficients.
			 */
			for _, descriptor := range descriptors {
				filterName := descriptor.Name
				wavePath := descriptor.Path
				dc := descriptor.Compensation
				dcFloat := float64(dc)
				compensation := 0.05 * dcFloat
				fac := math.Pow(10.0, compensation)
				waveBuffer, err := ioutil.ReadFile(wavePath)

				/*
				 * Check if file was read successfully.
				 */
				if err != nil {
					fmt.Printf("WARNING: During filter import: Could not read file '%s'. - Skipping.\n", wavePath)
				} else {
					waveFile, err := wave.FromBuffer(waveBuffer)

					/*
					 * Check if file was parsed successfully.
					 */
					if err != nil {
						fmt.Printf("WARNING: During filter import (file '%s'): %s\n", wavePath, err.Error())
					} else {
						channelCount := waveFile.ChannelCount()

						/*
						 * An FIR filter should have exactly one channel.
						 */
						if channelCount != CHANNEL_COUNT {
							fmt.Printf("WARNING: During filter import: File '%s' contains %d channels, expected: %d - Skipping.\n", wavePath, channelCount, CHANNEL_COUNT)
						} else {
							sampleRate := waveFile.SampleRate()
							channel, _ := waveFile.Channel(0)
							content := channel.Floats()

							/*
							 * Iterate over the supported sample rates.
							 */
							for _, targetSampleRate := range g_sampleRates {
								coefficients := resample.Time(content, sampleRate, targetSampleRate)

								/*
								 * Create impulse response structure.
								 */
								ir := impulseResponseStruct{
									name:             filterName,
									gainCompensation: fac,
									sampleRate:       targetSampleRate,
									data:             coefficients,
								}

								impulseResponseList = append(impulseResponseList, ir)
							}

						}

					}

				}

			}

			/*
			 * Create data structure for impulse responses.
			 */
			impulseResponses := impulseResponsesStruct{
				responses: impulseResponseList,
			}

			return &impulseResponses, nil
		}

	}

}

/*
 * Create an empty filter, which does not pass any signal.
 */
func Empty(sampleRate uint32) Filter {
	coeffs := make([]float64, 0)

	/*
	 * Create impulse response.
	 */
	ir := impulseResponseStruct{
		name:             "(EMPTY)",
		gainCompensation: 0.0,
		sampleRate:       sampleRate,
		data:             coeffs,
	}

	ft := fft.CreateFourierTransform()
	bufFilterC := make([]complex128, 0)
	bufFilteredC := make([]complex128, 0)
	bufInput := make([]float64, 0)
	bufInputC := make([]complex128, 0)
	bufOutput := make([]float64, 0)
	bufOutputC := make([]complex128, 0)
	bufTail := make([]float64, 0)

	/*
	 * Create a new filter.
	 */
	fltFilter := filterStruct{
		impulseResponse:     ir,
		fourierTransform:    ft,
		filterComplex:       bufFilterC,
		filteredComplex:     bufFilteredC,
		inputBuffer:         bufInput,
		inputBufferComplex:  bufInputC,
		outputBuffer:        bufOutput,
		outputBufferComplex: bufOutputC,
		tailBuffer:          bufTail,
	}

	return &fltFilter
}

/*
 * Creates a filter from a list of coefficients.
 */
func FromCoefficients(coeffs []float64, sampleRate uint32, name string) Filter {
	numCoeffs := len(coeffs)
	coeffsCopy := make([]float64, numCoeffs)
	copy(coeffsCopy, coeffs)

	/*
	 * Create impulse response.
	 */
	ir := impulseResponseStruct{
		name:             name,
		gainCompensation: 0.0,
		sampleRate:       sampleRate,
		data:             coeffsCopy,
	}

	ft := fft.CreateFourierTransform()
	bufFilterC := make([]complex128, 0)
	bufFilteredC := make([]complex128, 0)
	bufInput := make([]float64, 0)
	bufInputC := make([]complex128, 0)
	bufOutput := make([]float64, 0)
	bufOutputC := make([]complex128, 0)
	bufTail := make([]float64, 0)

	/*
	 * Create a new filter.
	 */
	fltFilter := filterStruct{
		impulseResponse:     ir,
		fourierTransform:    ft,
		filterComplex:       bufFilterC,
		filteredComplex:     bufFilteredC,
		inputBuffer:         bufInput,
		inputBufferComplex:  bufInputC,
		outputBuffer:        bufOutput,
		outputBufferComplex: bufOutputC,
		tailBuffer:          bufTail,
	}

	return &fltFilter
}

/*
 * Returns the supported sample rates.
 */
func SampleRates() []uint32 {
	numRates := len(g_sampleRates)
	sampleRates := make([]uint32, numRates)
	copy(sampleRates, g_sampleRates)
	return sampleRates
}
