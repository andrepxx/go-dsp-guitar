package resample

import (
	"math"
	"testing"
)

/*
 * Compare two real-valued slices to check whether their components are close.
 */
func areSlicesClose(a []float64, b []float64) (bool, []float64) {

	/*
	 * Check whether the two slices are of the same size.
	 */
	if len(a) != len(b) {
		return false, nil
	} else {
		c := true
		n := len(a)
		diffs := make([]float64, n)

		/*
		 * Iterate over the arrays to compare values.
		 */
		for i, elem := range a {
			diff := elem - b[i]
			diffAbs := math.Abs(diff)

			/*
			 * Check if we found a significant difference.
			 */
			if diffAbs > 0.00000001 {
				c = false
			}

			diffs[i] = diff
		}

		return c, diffs
	}

}

/*
 * Perform a unit test on the Lanczos algorithm for time series data.
 */
func TestTimeSeries(t *testing.T) {

	/*
	 * Input vectors.
	 */
	in := [][]float64{
		[]float64{0.87622011, 0.41920066, 0.56935138, 0.56090797, 0.0485888, 0.89798242, 0.94420837, 0.89861948},
	}

	/*
	 * Expected output vectors for upsampling.
	 */
	outExpectedUp := [][]float64{
		[]float64{0.87622011, 0.72424457, 0.41920066, 0.40800042, 0.56935138, 0.66706275, 0.56090797, 0.20545441, 0.0485888, 0.40780951, 0.89798242, 1.00559434, 0.94420837, 1.00017368, 0.89861948},
	}

	/*
	 * Expected output vectors for downsampling.
	 */
	outExpectedDown := [][]float64{
		[]float64{0.87622011, 0.61602851, 0.25912048},
	}

	/*
	 * Test with each input vector.
	 */
	for i, currentIn := range in {
		expectedUp := outExpectedUp[i]
		expectedDown := outExpectedDown[i]
		currentResultUp := Time(currentIn, 96000, 192000)
		currentResultDown := Time(currentIn, 96000, 44100)
		okUp, diffUp := areSlicesClose(currentResultUp, expectedUp)

		/*
		 * Verify components of upsampled vector.
		 */
		if !okUp {
			t.Errorf("Upsampling vector number %d: Result is incorrect. Expected %v, got %v, difference: %v", i, expectedUp, currentResultUp, diffUp)
		}

		okDown, diffDown := areSlicesClose(currentResultDown, expectedDown)

		/*
		 * Verify components of downsampled vector.
		 */
		if !okDown {
			t.Errorf("Downsampling vector number %d: Result is incorrect. Expected %v, got %v, difference: %v", i, expectedDown, currentResultDown, diffDown)
		}

	}

}

/*
 * Perform a unit test on the Lanczos algorithm for frequency series data.
 */
func TestFrequencySeries(t *testing.T) {

	/*
	 * Complex input vectors.
	 */
	in := [][]complex128{
		[]complex128{
			complex(0.34233881, 0.25689662),
			complex(0.04731972, 0.70090472),
			complex(0.6126194, 0.21446363),
			complex(0.4184522, 0.44984173),
			complex(0.58391517, 0.93459223),
			complex(0.52775765, 0.05379716),
			complex(0.13449256, 0.70627374),
			complex(0.05077271, 0.49363423),
		},
	}

	/*
	 * Real part of expected output vectors for downsampling.
	 */
	outExpectedReal := [][]float64{
		[]float64{0.34233881, 0.6126194, 0.58391517, 0.13449256},
	}

	/*
	 * Imaginary part of expected output vectors for downsampling.
	 */
	outExpectedImag := [][]float64{
		[]float64{0.25689662, 0.21446363, 0.93459223, 0.70627374},
	}

	/*
	 * Test with each input vector.
	 */
	for i, currentIn := range in {
		expectedReal := outExpectedReal[i]
		expectedImag := outExpectedImag[i]
		currentResult := Frequency(currentIn, 4)
		nResult := len(currentResult)
		currentResultReal := make([]float64, nResult)
		currentResultImag := make([]float64, nResult)

		/*
		 * Extract real and imaginary part from result.
		 */
		for i, elem := range currentResult {
			currentResultReal[i] = real(elem)
			currentResultImag[i] = imag(elem)
		}

		okReal, diffReal := areSlicesClose(currentResultReal, expectedReal)

		/*
		 * Verify components of upsampled vector.
		 */
		if !okReal {
			t.Errorf("Real vector %d: Result is incorrect. Expected %v, got %v, difference: %v", i, expectedReal, currentResultReal, diffReal)
		}

		okImag, diffImag := areSlicesClose(currentResultImag, expectedImag)

		/*
		 * Verify components of downsampled vector.
		 */
		if !okImag {
			t.Errorf("Imaginary vector %d: Result is incorrect. Expected %v, got %v, difference: %v", i, expectedImag, currentResultImag, diffImag)
		}

	}

}
