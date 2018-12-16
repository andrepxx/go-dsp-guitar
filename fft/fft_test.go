package fft

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
 * Compare two complex-valued slices to check whether their components are equal.
 */
func areSlicesEqual(a []complex128, b []complex128) bool {

	/*
	 * Check whether the two slices are of the same size.
	 */
	if len(a) != len(b) {
		return false
	} else {
		c := true

		/*
		 * Iterate over the arrays to compare values.
		 */
		for i, elem := range a {
			c = c && (elem == b[i])
		}

		return c
	}

}

/*
 * Perform a unit test on the power-of-two function.
 */
func TestNextPowerOfTwo(t *testing.T) {

	/*
	 * Input values.
	 */
	in := []uint64{
		6493615572477977987,
		183778605738611348,
		1211049956568627,
		877784,
		65537,
		65536,
		63128,
		255,
		2,
		1,
	}

	/*
	 * Next higher powers of two.
	 */
	powers := []uint64{
		9223372036854775808,
		288230376151711744,
		2251799813685248,
		1048576,
		131072,
		65536,
		65536,
		256,
		2,
		1,
	}

	/*
	 * Exponents.
	 */
	exponents := []uint32{
		63,
		58,
		51,
		20,
		17,
		16,
		16,
		8,
		1,
		0,
	}

	/*
	 * Test the function for each input value.
	 */
	for i, val := range in {
		expectedP := powers[i]
		expectedE := exponents[i]
		p, e := NextPowerOfTwo(val)

		/*
		 * Check if we got the expected power.
		 */
		if p != expectedP {
			t.Errorf("Power of two number %d: Resulting power is incorrect. Expected %d (%x), got %d (%x).", i, expectedP, expectedP, p, p)
		}

		/*
		 * Check if we got the expected exponent.
		 */
		if e != expectedE {
			t.Errorf("Power of two number %d: Resulting exponent is incorrect. Expected %d (%x), got %d (%x).", i, expectedE, expectedE, e, e)
		}

	}

}

/*
 * Perform a unit test on the zeroes a complex-valued vector.
 */
func TestZeroComplex(t *testing.T) {
	sizes := []int{1, 2, 4, 8, 15, 16}

	/*
	 * Create buffers of different size.
	 */
	for _, size := range sizes {
		ones := make([]complex128, size)

		/*
		 * Initialize both real and imaginary part to one.
		 */
		for i, _ := range ones {
			ones[i] = complex(1.0, 1.0)
		}

		ZeroComplex(ones)

		/*
		 * Verify that all elements are now zeroed.
		 */
		for i, elem := range ones {
			re := real(elem)
			ie := imag(elem)

			/*
			 * If either real or imaginary part is non-zero, fail.
			 */
			if re != 0.0 || ie != 0.0 {
				t.Errorf("Failed to zero complex-valued buffer of size %d. Element %d is non-zero.", size, i)
			}

		}

	}

}

/*
 * Perform a unit test on the function that zeroes a real-valued vector.
 */
func TestZeroFloat(t *testing.T) {
	sizes := []int{1, 2, 4, 8, 15, 16}

	/*
	 * Create buffers of different size.
	 */
	for _, size := range sizes {
		ones := make([]float64, size)

		/*
		 * Initialize each element to one.
		 */
		for i, _ := range ones {
			ones[i] = 1.0
		}

		ZeroFloat(ones)

		/*
		 * Verify that all elements are now zeroed.
		 */
		for i, elem := range ones {

			/*
			 * If element is non-zero, fail.
			 */
			if elem != 0.0 {
				t.Errorf("Failed to zero real-valued buffer of size %d. Element %d is non-zero.", size, i)
			}

		}

	}

}

/*
 * Perform a unit test on the real-valued FFT.
 */
func TestRealFFT(t *testing.T) {

	/*
	 * Input vectors.
	 */
	in := [][]float64{
		[]float64{0.0, 1.0, 0.0, 0.0},
		[]float64{1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{1.0, 2.0, 3.0, 4.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
		[]float64{1.0, -1.0, 1.0, -1.0, 1.0, -1.0, 1.0, -1.0},
		[]float64{0.93990505, 0.20043027, 0.24328743, 0.39466036, 0.62847371, 0.29570877, 0.30114516, 0.7491788},
	}

	/*
	 * Real components of expected output vectors.
	 */
	outRealExpected := [][]float64{
		[]float64{1.0, 0.0, -1.0, 0.0},
		[]float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
		[]float64{1.0, 0.70710678, 0.0, -0.70710678, -1.0, -0.70710678, 0.0, 0.70710678},
		[]float64{10.0, -0.41421356, -2.00000000, 2.41421356, -2.0, 2.41421356, -2.0, -0.41421356},
		[]float64{36.0, -4.0, -4.0, -4.0, -4.0, -4.0, -4.0, -4.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 8.0, 0.0, 0.0, 0.0},
		[]float64{3.75278955, 0.49474166, 1.02394617, 0.12812102, 0.47283315, 0.12812102, 1.02394617, 0.49474166},
	}

	/*
	 * Imaginary components of expected output vectors.
	 */
	outImagExpected := [][]float64{
		[]float64{0.0, -1.0, 0.0, 1.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, -0.70710678, -1.0, -0.70710678, 0.0, 0.70710678, 1.0, 0.70710678},
		[]float64{0.0, -7.24264069, 2.0, -1.24264069, 0.0, 1.24264069, -2.0, 7.24264069},
		[]float64{0.0, 9.65685425, 4.0, 1.65685425, 0.0, -1.65685425, -4., -9.65685425},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.3759122, 0.64770012, 0.26019674, 0.0, -0.26019674, -0.64770012, -0.3759122},
	}

	/*
	 * Test with each input vector.
	 */
	for i, currentIn := range in {
		n := len(currentIn)
		expectedReal := outRealExpected[i]
		nReal := len(expectedReal)
		expectedImag := outImagExpected[i]
		nImag := len(expectedImag)

		/*
		 * Verify that expected result vectors are of correct size.
		 */
		if nReal != n {
			t.Errorf("Expected real result vector %d is of incorrect size: Expected %d, got %d.", i, n, nReal)
		} else if nImag != n {
			t.Errorf("Expected imaginary result vector %d is of incorrect size: Expected %d, got %d.", i, n, nImag)
		} else {
			currentResult := make([]complex128, nReal)
			err := RealFourier(currentIn, currentResult, SCALING_DEFAULT)

			/*
			 * Check if forward transform was calculated successfully.
			 */
			if err != nil {
				msg := err.Error()
				t.Errorf("Failed to calculate real FFT: %s", msg)
			} else {
				currentResultReal := make([]float64, nReal)
				currentResultImag := make([]float64, nImag)

				/*
				 * Extract real and imaginary components from result vector.
				 */
				for i, elem := range currentResult {
					currentResultReal[i] = real(elem)
					currentResultImag[i] = imag(elem)
				}

				okReal, diffReal := areSlicesClose(currentResultReal, expectedReal)

				/*
				 * Verify real components of result vector.
				 */
				if !okReal {
					t.Errorf("Real FFT number %d: Real part of result is incorrect. Expected %v, got %v, difference: %v", i, expectedReal, currentResultReal, diffReal)
				}

				okImag, diffImag := areSlicesClose(currentResultImag, expectedImag)

				/*
				 * Verify imaginary components of result vector.
				 */
				if !okImag {
					t.Errorf("Real FFT number %d: Imaginary part of result is incorrect. Expected %v, got %v, difference: %v", i, expectedImag, currentResultImag, diffImag)
				}

			}

			currentInverse := make([]float64, nReal)
			err = RealInverseFourier(currentResult, currentInverse, SCALING_DEFAULT)

			/*
			 * Check if inverse transform was calculated successfully.
			 */
			if err != nil {
				msg := err.Error()
				t.Errorf("Failed to calculate real IFFT: %s", msg)
			} else {
				okInverse, diffInverse := areSlicesClose(currentInverse, currentIn)

				/*
				 * Verify components of IFFT result vector.
				 */
				if !okInverse {
					t.Errorf("Real IFFT number %d: Result is incorrect. Expected %v, got %v, difference: %v", i, currentIn, currentInverse, diffInverse)
				}

			}

		}

	}

}

/*
 * Perform a unit test on the complex-valued FFT.
 */
func TestComplexFFT(t *testing.T) {

	/*
	 * Real components fo the input vectors.
	 */
	inReal := [][]float64{
		[]float64{0.0, 1.0, 0.0, 0.0},
		[]float64{1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{1.0, 2.0, 3.0, 4.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
		[]float64{1.0, -1.0, 1.0, -1.0, 1.0, -1.0, 1.0, -1.0},
		[]float64{0.93811391, 0.12498467, 0.65156107, 0.68689968, 0.04341771, 0.29019219, 0.89338032, 0.44420547},
	}

	/*
	 * Imaginary components fo the input vectors.
	 */
	inImag := [][]float64{
		[]float64{0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.00579331, 0.57801897, 0.69192584, 0.60747351, 0.75338567, 0.24053831, 0.12623075, 0.01731368},
	}

	/*
	 * Real components of expected output vectors.
	 */
	outRealExpected := [][]float64{
		[]float64{1.0, 0.0, -1.0, 0.0},
		[]float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
		[]float64{1.0, 0.70710678, 0.0, -0.70710678, -1.0, -0.70710678, 0.0, 0.70710678},
		[]float64{10.0, -0.41421356, -2.00000000, 2.41421356, -2.0, 2.41421356, -2.0, -0.41421356},
		[]float64{36.0, -4.0, -4.0, -4.0, -4.0, -4.0, -4.0, -4.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 8.0, 0.0, 0.0, 0.0},
		[]float64{4.07275502, 1.82790209, -0.36963968, 1.27337207, 0.98019100, 1.09288049, -0.75717986, -0.61536985},
	}

	/*
	 * Imaginary components of expected output vectors.
	 */
	outImagExpected := [][]float64{
		[]float64{0.0, -1.0, 0.0, 1.0},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{0.0, -0.70710678, -1.0, -0.70710678, 0.0, 0.70710678, 1.0, 0.70710678},
		[]float64{0.0, -7.24264069, 2.0, -1.24264069, 0.0, 1.24264069, -2.0, 7.24264069},
		[]float64{0.0, 9.65685425, 4.0, 1.65685425, 0.0, -1.65685425, -4., -9.65685425},
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		[]float64{3.02068004, -0.73923563, 0.65695068, -0.86553182, 0.1339911, -0.27231059, -0.7749059, -1.1132914},
	}

	/*
	 * Possible modes of operation.
	 */
	modes := []int{
		MODE_STANDARD,
		MODE_INPLACE,
	}

	/*
	 * Test with each mode of operation.
	 */
	for _, mode := range modes {

		/*
		 * Test with each input vector.
		 */
		for i, currentInReal := range inReal {
			n := len(currentInReal)
			currentInImag := inImag[i]
			m := len(currentInImag)

			/*
			 * Verify that the input vectors are of correct size.
			 */
			if n != m {
				t.Errorf("Components of vector %d are of unequal size: %d real components and %d imaginary components.", i, n, m)
			} else {
				currentComplexIn := make([]complex128, n)

				/*
				 * Form a complex vector out of real and imaginary components.
				 */
				for j, cr := range currentInReal {
					ci := currentInImag[j]
					currentComplexIn[j] = complex(cr, ci)
				}

				expectedReal := outRealExpected[i]
				nReal := len(expectedReal)
				expectedImag := outImagExpected[i]
				nImag := len(expectedImag)

				/*
				 * Verify that expected result vectors are of correct size.
				 */
				if nReal != n {
					t.Errorf("Expected real result vector %d is of incorrect size: Expected %d, got %d.", i, n, nReal)
				} else if nImag != n {
					t.Errorf("Expected imaginary result vector %d is of incorrect size: Expected %d, got %d.", i, n, nImag)
				} else {
					currentResult := make([]complex128, n)
					copy(currentResult, currentComplexIn)
					currentResult = Fourier(currentResult, SCALING_DEFAULT, mode)
					currentResultReal := make([]float64, nReal)
					currentResultImag := make([]float64, nImag)

					/*
					 * Extract real and imaginary components from result vector.
					 */
					for i, elem := range currentResult {
						currentResultReal[i] = real(elem)
						currentResultImag[i] = imag(elem)
					}

					okReal, diffReal := areSlicesClose(currentResultReal, expectedReal)

					/*
					 * Verify real components of result vector.
					 */
					if !okReal {
						t.Errorf("Complex FFT number %d (mode %d): Real part of result is incorrect. Expected %v, got %v, difference: %v", i, mode, expectedReal, currentResultReal, diffReal)
					}

					okImag, diffImag := areSlicesClose(currentResultImag, expectedImag)

					/*
					 * Verify imaginary components of result vector.
					 */
					if !okImag {
						t.Errorf("Complex FFT number %d (mode %d): Imaginary part of result is incorrect. Expected %v, got %v, difference: %v", i, mode, expectedImag, currentResultImag, diffImag)
					}

					currentInverse := make([]complex128, n)
					copy(currentInverse, currentResult)
					currentInverse = InverseFourier(currentInverse, SCALING_DEFAULT, mode)
					currentInverseReal := make([]float64, nReal)
					currentInverseImag := make([]float64, nImag)

					/*
					 * Extract real and imaginary components from inverse vector.
					 */
					for i, elem := range currentInverse {
						currentInverseReal[i] = real(elem)
						currentInverseImag[i] = imag(elem)
					}

					okInverseReal, diffInverseReal := areSlicesClose(currentInverseReal, currentInReal)

					/*
					 * Verify real components of IFFT result vector.
					 */
					if !okInverseReal {
						t.Errorf("Complex IFFT number %d (mode %d): Real part of result is incorrect. Expected %v, got %v, difference: %v", i, mode, currentInReal, currentInverseReal, diffInverseReal)
					}

					okInverseImag, diffInverseImag := areSlicesClose(currentInverseImag, currentInImag)

					/*
					 * Verify real components of IFFT result vector.
					 */
					if !okInverseImag {
						t.Errorf("Complex IFFT number %d (mode %d): Imaginary part of result is incorrect. Expected %v, got %v, difference: %v", i, mode, currentInImag, currentInverseImag, diffInverseImag)
					}

				}

			}

		}

	}

}

/*
 * Test (real) FFT with orthonormal scaling.
 */
func TestOrthonormalScaling(t *testing.T) {

	/*
	 * Real input vector.
	 */
	in := []float64{
		0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	}

	/*
	 * Real part of expected output vector.
	 */
	expectedReal := []float64{
		0.35355339, 0.25, 0.0, -0.25, -0.35355339, -0.25, 0.0, 0.25,
	}

	/*
	 * Imaginary part of expected output vector.
	 */
	expectedImag := []float64{
		0.0, -0.25, -0.35355339, -0.25, 0.0, 0.25, 0.35355339, 0.25,
	}

	n := len(in)
	result := make([]complex128, n)
	err := RealFourier(in, result, SCALING_ORTHONORMAL)

	/*
	 * Check if forward transform was calculated successfully.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to calculate real orthonormal FFT: %s", msg)
	} else {
		resultReal := make([]float64, n)
		resultImag := make([]float64, n)

		/*
		 * Extract real and imaginary components from result vector.
		 */
		for i, elem := range result {
			resultReal[i] = real(elem)
			resultImag[i] = imag(elem)
		}

		okReal, diffReal := areSlicesClose(resultReal, expectedReal)

		/*
		 * Verify real components of result vector.
		 */
		if !okReal {
			t.Errorf("Real orthonormal FFT: Real part of result is incorrect. Expected %v, got %v, difference: %v", expectedReal, resultReal, diffReal)
		}

		okImag, diffImag := areSlicesClose(resultImag, expectedImag)

		/*
		 * Verify imaginary components of result vector.
		 */
		if !okImag {
			t.Errorf("Real orthonormal FFT: Imaginary part of result is incorrect. Expected %v, got %v, difference: %v", expectedImag, resultImag, diffImag)
		}

	}

	inverse := make([]float64, n)
	err = RealInverseFourier(result, inverse, SCALING_ORTHONORMAL)

	/*
	 * Check if inverse transform was calculated successfully.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to calculate real orthonormal IFFT: %s", msg)
	} else {
		okInverse, diffInverse := areSlicesClose(inverse, in)

		/*
		 * Verify components of IFFT result vector.
		 */
		if !okInverse {
			t.Errorf("Real orthonormal IFFT: Result is incorrect. Expected %v, got %v, difference: %v", in, inverse, diffInverse)
		}

	}

}

/*
 * Test the edge-case of an input vector containing only a single element.
 */
func TestSingleElementFFT(t *testing.T) {

	/*
	 * Real input.
	 */
	inReal := []float64{
		3.14,
	}

	/*
	 * Complex input.
	 */
	inComplex := []complex128{
		complex(3.14, 0.0),
	}

	outReal := make([]float64, 1)
	outComplex := make([]complex128, 1)
	err := RealFourier(inReal, outComplex, SCALING_DEFAULT)

	/*
	 * Check if forward transform was calculated successfully.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Single-element real FFT failed: %s", msg)
	} else {

		/*
		 * Check if we got the expected result for the forward transform.
		 */
		if real(outComplex[0]) != 3.14 || imag(outComplex[0]) != 0.0 {
			t.Errorf("Single-element real FFT did not return expected result.")
		}

		err = RealInverseFourier(inComplex, outReal, SCALING_DEFAULT)

		/*
		 * Check if inverse transform was calculated successfully.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Single-element real IFFT failed: %s", msg)
		} else {

			/*
			 * Check if we got the expected result for the inverse transform.
			 */
			if outReal[0] != 3.14 {
				t.Errorf("Single-element real IFFT did not return expected result.")
			}

		}

	}

}

/*
 * Test cases where the transforms should fail.
 */
func TestFailureCases(t *testing.T) {
	rThree := make([]float64, 3)
	rEight := make([]float64, 8)
	cThree := make([]complex128, 3)
	cFour := make([]complex128, 4)
	err := RealFourier(rThree, cThree, SCALING_DEFAULT)

	/*
	 * Verify that the transform failed.
	 */
	if err == nil {
		t.Errorf("Real FFT of size three did not fail!")
	}

	err = RealInverseFourier(cThree, rThree, SCALING_DEFAULT)

	/*
	 * Verify that the transform failed.
	 */
	if err == nil {
		t.Errorf("Real IFFT of size three did not fail!")
	}

	err = RealFourier(rEight, cFour, SCALING_DEFAULT)

	/*
	 * Verify that the transform failed.
	 */
	if err == nil {
		t.Errorf("Real FFT of unequal size did not fail!")
	}

	err = RealInverseFourier(cFour, rEight, SCALING_DEFAULT)

	/*
	 * Verify that the transform failed.
	 */
	if err == nil {
		t.Errorf("Real IFFT of unequal size did not fail!")
	}

}

/*
 * Test spectral shifting functions.
 */
func TestShift(t *testing.T) {

	/*
	 * Input vector with even number of elements.
	 */
	inEven := []complex128{
		complex(1.0, 2.0),
		complex(3.0, 4.0),
		complex(5.0, 6.0),
		complex(7.0, 8.0),
	}

	/*
	 * Expected output vector with even number of elements.
	 */
	outEven := []complex128{
		complex(5.0, 6.0),
		complex(7.0, 8.0),
		complex(1.0, 2.0),
		complex(3.0, 4.0),
	}

	/*
	 * Input vector with odd number of elements.
	 */
	inOdd := []complex128{
		complex(1.0, 2.0),
		complex(3.0, 4.0),
		complex(5.0, 6.0),
		complex(7.0, 8.0),
		complex(9.0, 10.0),
	}

	/*
	 * Expected output vector with even number of elements.
	 */
	outOdd := []complex128{
		complex(7.0, 8.0),
		complex(9.0, 10.0),
		complex(1.0, 2.0),
		complex(3.0, 4.0),
		complex(5.0, 6.0),
	}

	nEven := len(inEven)
	nOdd := len(inOdd)
	shiftEven := make([]complex128, nEven)
	shiftOdd := make([]complex128, nOdd)
	copy(shiftEven, inEven)
	copy(shiftOdd, inOdd)
	Shift(shiftEven, false)
	Shift(shiftOdd, false)
	evenCorrect := areSlicesEqual(shiftEven, outEven)

	/*
	 * Check if array with even element-count was permuted correctly.
	 */
	if !evenCorrect {
		t.Errorf("Even-valued array was not permuted correctly. Expected %v, got %v.", outEven, shiftEven)
	}

	oddCorrect := areSlicesEqual(shiftOdd, outOdd)

	/*
	 * Check if array with odd element-count was permuted correctly.
	 */
	if !oddCorrect {
		t.Errorf("Odd-valued array was not permuted correctly. Expected %v, got %v.", outOdd, shiftOdd)
	}

	Shift(shiftEven, true)
	Shift(shiftOdd, true)
	evenCorrect = areSlicesEqual(shiftEven, inEven)

	/*
	 * Check if array with even element-count was inverse-permuted correctly.
	 */
	if !evenCorrect {
		t.Errorf("Even-valued array was not inverse-permuted correctly. Expected %v, got %v.", inEven, shiftEven)
	}

	oddCorrect = areSlicesEqual(shiftOdd, inOdd)

	/*
	 * Check if array with even element-count was inverse-permuted correctly.
	 */
	if !evenCorrect {
		t.Errorf("Odd-valued array was not inverse-permuted correctly. Expected %v, got %v.", inOdd, shiftOdd)
	}

}
