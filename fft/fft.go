package fft

import (
	"fmt"
	"math"
	"math/bits"
	"math/cmplx"
	"sync"
)

/*
 * Global constants.
 */
const (
	SCALING_DEFAULT = iota
	SCALING_ORTHONORMAL
	MODE_STANDARD
	MODE_INPLACE
)

/*
 * Mathematical constants
 */
const (
	MATH_MINUS_TWO_PI = -2.0 * math.Pi
)

/*
 * Global variables.
 */
var g_mutex sync.Mutex
var g_coefficientsLarge map[int][]complex128
var g_coefficientsSmall []complex128
var g_permutationLarge map[int][]int
var g_permutationSmall []int
var g_scrapspace []complex128

/*
 * Generates Fourier coefficients for n = 1, 2, 4, 8, ..., 8192.
 */
func generateFourierCoefficients() []complex128 {
	coefficients := make([]complex128, 16384)
	coefficients[0] = complex(0.0, 0.0)

	/*
	 * Generate coefficients for n = 2^k with k = [0, 13].
	 */
	for k := uint(0); k <= 13; k++ {
		n := 1 << k
		nFloat := float64(n)

		/*
		 * Generate n coefficients.
		 */
		for i := 0; i < n; i++ {
			idx := n + i
			iFloat := float64(i)
			argImag := (MATH_MINUS_TWO_PI * iFloat) / nFloat
			arg := complex(0.0, argImag)
			coefficients[idx] = cmplx.Exp(arg)
		}

	}

	return coefficients
}

/*
 * Generates permutation coefficients for n = 1, 2, 4, 8, ..., 8192.
 */
func generatePermutationCoefficients() []int {
	coefficients := make([]int, 16384)
	coefficients[0] = 0

	/*
	 * Generate coefficients for n = 2^p with p = [0, 13].
	 */
	for p := 0; p <= 13; p++ {
		shiftP := uint(p)
		n := 1 << shiftP
		coefficients[n] = 0

		/*
		 * Do this for every power of two.
		 */
		for k := 0; k < p; k++ {
			shiftM := uint(k)
			m := 1 << shiftM

			/*
			 * Copy the coefficients, shifted by one place, then increment
			 * them by one.
			 */
			for i := 0; i < m; i++ {
				idxA := n + i
				idxB := idxA + m
				value := coefficients[idxA]
				value <<= 1
				coefficients[idxA] = value
				coefficients[idxB] = value + 1
			}

		}

	}

	return coefficients
}

/*
 * Returns the Fourier coefficients for a Fourier transform of the specified size.
 */
func fourierCoefficients(n int) []complex128 {

	/*
	 * Ensure that the number of coefficients is positive, then fetch them either
	 * from a slice or generate them and store them in a map.
	 */
	if n < 0 {
		return nil
	} else if n <= 8192 {
		uBound := n << 1
		return g_coefficientsSmall[n:uBound]
	} else {
		g_mutex.Lock()
		coefficients, ok := g_coefficientsLarge[n]

		/*
		 * If coefficients aren't already calculated, calculate them now.
		 */
		if !ok {
			coefficients = make([]complex128, n)
			nFloat := float64(n)

			/*
			 * Calculate the Fourier coefficients.
			 */
			for j := 0; j < n; j++ {
				jFloat := float64(j)
				argImag := (MATH_MINUS_TWO_PI * jFloat) / nFloat
				arg := complex(0.0, argImag)
				coefficients[j] = cmplx.Exp(arg)
			}

			g_coefficientsLarge[n] = coefficients
		}

		g_mutex.Unlock()
		return coefficients
	}

}

/*
 * Returns the permutation coefficients for an in-place Fourier transform of the
 * specified size.
 */
func permutationCoefficients(n int) []int {

	/*
	 * Ensure that the number of coefficients is positive, then fetch them either
	 * from a slice or generate them and store them in a map.
	 */
	if n < 0 {
		return nil
	} else if n <= 8192 {
		uBound := n << 1
		return g_permutationSmall[n:uBound]
	} else {
		g_mutex.Lock()
		coefficients, ok := g_permutationLarge[n]

		/*
		 * If coefficients aren't already calculated, calculate them now.
		 */
		if !ok {
			n64 := uint64(n)
			num, p := NextPowerOfTwo(n64)
			coefficients = make([]int, num)
			coefficients[0] = 0

			/*
			 * Do this for every power of two.
			 */
			for i := uint32(0); i < p; i++ {
				m := 1 << i

				/*
				 * Copy the coefficients, shifted by one place, then
				 * increment them by one.
				 */
				for j := 0; j < m; j++ {
					value := coefficients[j]
					value <<= 1
					coefficients[j] = value
					idx := j + m
					coefficients[idx] = value + 1
				}

			}

			g_permutationLarge[n] = coefficients
		}

		g_mutex.Unlock()
		return coefficients
	}

}

/*
 * Compute the fast Fourier transform using the recursive Cooley-Tukey algorithm.
 */
func cooleyTukey(vec []complex128) []complex128 {
	n := len(vec)

	/*
	 * Abort recursion when only a single element is left.
	 */
	if n <= 1 {
		return vec
	} else {
		nHalf := n / 2
		even := make([]complex128, nHalf)
		odd := make([]complex128, nHalf)
		result := make([]complex128, n)

		/*
		 * Split vector into even and odd half.
		 */
		for i := 0; i < nHalf; i++ {
			idxEven := i << 1
			idxOdd := idxEven + 1
			even[i] = vec[idxEven]
			odd[i] = vec[idxOdd]
		}

		lower := cooleyTukey(even)
		upper := cooleyTukey(odd)
		coefficients := fourierCoefficients(n)

		/*
		 * Perform the "twiddling".
		 */
		for i, elem := range lower {
			product := coefficients[i] * upper[i]
			lower[i] = elem + product
			upper[i] = elem - product
		}

		copy(result[0:nHalf], lower)
		copy(result[nHalf:n], upper)
		return result
	}

}

/*
 * Perform the Fourier input permutation on a vector.
 */
func permute(vec []complex128) {
	n := len(vec)
	coeff := permutationCoefficients(n)
	g_mutex.Lock()

	/*
	 * Check if size for scrapspace is sufficient.
	 */
	if g_scrapspace == nil || len(g_scrapspace) < n {
		g_scrapspace = make([]complex128, n)
	}

	copy(g_scrapspace, vec)

	/*
	 * Permute the elements.
	 */
	for i := 0; i < n; i++ {
		idx := coeff[i]
		vec[i] = g_scrapspace[idx]
	}

	g_mutex.Unlock()
}

/*
 * Compute the fast Fourier transform using an (unnamed?) in-place algorithm.
 */
func inplaceTransform(vec []complex128) {
	permute(vec)
	n := len(vec)
	coeffs := fourierCoefficients(n)
	size := 1
	stride := n
	n64 := uint64(n)
	npp := n64 + 1
	_, p := NextPowerOfTwo(npp)
	pmm := int(p - 1)

	/*
	 * Fourier rounds.
	 */
	for i := 1; i <= pmm; i++ {
		size <<= 1
		stride >>= 1
		blocks := n / size // The number of blocks.

		/*
		 * Process each block.
		 */
		for j := 0; j < blocks; j++ {
			halfBlocks := blocks << 1
			half := n / halfBlocks // The length of a half-block.
			dj := j << 1
			offset := dj * half // The offset into the current block.

			/*
			 * Perform the butterfly operations.
			 */
			for k := 0; k < half; k++ {
				i := k + offset
				j := i + half
				vi := vec[i]
				vj := vec[j]
				l := k * stride
				m := half * stride
				n := l + m
				cl := coeffs[l]
				cn := coeffs[n]
				left := vi + (cl * vj)
				right := vi + (cn * vj)
				vec[i] = left
				vec[j] = right
			}

		}

	}

}

/*
 * Initialize the computation of a Fourier transform.
 */
func initialize() {
	g_mutex.Lock()

	/*
	 * Generate the global Fourier coefficient slice.
	 */
	if g_coefficientsSmall == nil {
		g_coefficientsSmall = generateFourierCoefficients()
	}

	/*
	 * Initialize the global Fourier coefficient map.
	 */
	if g_coefficientsLarge == nil {
		g_coefficientsLarge = make(map[int][]complex128)
	}

	/*
	 * Initialize the global permutation coefficient slice.
	 */
	if g_permutationSmall == nil {
		g_permutationSmall = generatePermutationCoefficients()
	}

	/*
	 * Initialize the global permutation coefficient map.
	 */
	if g_permutationLarge == nil {
		g_permutationLarge = make(map[int][]int)
	}

	g_mutex.Unlock()
}

/*
 * Swap the real and imaginary parts of a complex-valued vector and return the new
 * vector.
 */
func swapComplex(vec []complex128) []complex128 {
	n := len(vec)
	result := make([]complex128, n)

	/*
	 * Swap real and imaginary part for each element of the vector.
	 */
	for i, elem := range vec {
		elemReal := real(elem)
		elemImag := imag(elem)
		result[i] = complex(elemImag, elemReal)
	}

	return result
}

/*
 * Swap two elements in a complex vector.
 */
func swapComplexElements(vec []complex128, i int, j int) {
	tmp := vec[i]
	vec[i] = vec[j]
	vec[j] = tmp
}

/*
 * Swap the real and imaginary parts of a complex-valued vector in-place.
 */
func swapComplexInPlace(vec []complex128) {

	/*
	 * Swap real and imaginary part for each element of the vector.
	 */
	for i, elem := range vec {
		elemReal := real(elem)
		elemImag := imag(elem)
		result := complex(elemImag, elemReal)
		vec[i] = result
	}

}

/*
 * Find the next higher power of two.
 *
 * If value is already a power of two, the same value is returned.
 */
func NextPowerOfTwo(value uint64) (uint64, uint32) {
	digit := bits.Len64(value)
	digit32 := uint32(digit)
	exp := digit32 - 1
	pow := uint64(1)
	pow <<= exp

	/*
	 * If we are still below the threshold, we need an extra bit.
	 */
	if pow < value {
		exp++
		pow <<= 1
	}

	return pow, exp
}

/*
 * Write zeroes to a complex-valued buffer.
 */
func ZeroComplex(buffer []complex128) {

	/*
	 * Iterate over the buffer to zero it.
	 */
	for i, _ := range buffer {
		buffer[i] = complex(0.0, 0.0)
	}

}

/*
 * Write zeroes to a floating-point buffer.
 */
func ZeroFloat(buffer []float64) {

	/*
	 * Iterate over the buffer to zero it.
	 */
	for i, _ := range buffer {
		buffer[i] = float64(0.0)
	}

}

/*
 * Calculates the Fourier transform of a vector.
 */
func Fourier(vec []complex128, scaling int, mode int) []complex128 {
	initialize()
	result := vec

	/*
	 * Decide on which mode to operate.
	 */
	switch mode {

	/*
	 * Standard mode - copies data elements, slower.
	 */
	case MODE_STANDARD:
		result = cooleyTukey(vec)

	/*
	 * In-place mode - avoids copies of data elements, faster.
	 */
	case MODE_INPLACE:
		inplaceTransform(result)

	/*
	 * This should never happen.
	 */
	default:
		result = nil
	}

	/*
	 * Check if we should apply orthonormal scaling.
	 */
	if scaling == SCALING_ORTHONORMAL {

		/*
		 * Make sure that we got a result.
		 */
		if result != nil {
			n := len(vec)
			nFloat := float64(n)
			sqrtN := math.Sqrt(nFloat)
			r := 1.0 / sqrtN
			fac := complex(r, 0.0)

			/*
			 * Scale the result vector.
			 */
			for i := 0; i < n; i++ {
				result[i] *= fac
			}

		}

	}

	return result
}

/*
 * Calculates the inverse Fourier transform of a vector.
 */
func InverseFourier(vec []complex128, scaling int, mode int) []complex128 {
	initialize()
	n := len(vec)
	nFloat := float64(n)
	r := float64(0.0)

	/*
	 * Check which kind of scaling should be applied.
	 */
	switch scaling {
	case SCALING_DEFAULT:
		r = 1.0 / nFloat
		break
	case SCALING_ORTHONORMAL:
		sqrtN := math.Sqrt(nFloat)
		r = 1.0 / sqrtN
		break
	}

	scalingFac := complex(r, 0.0)

	/*
	 * Decide on which mode to operate.
	 */
	switch mode {

	/*
	 * Standard mode - copies data elements, slower.
	 */
	case MODE_STANDARD:
		swapped := swapComplex(vec)
		swappedResult := cooleyTukey(swapped)
		result := swapComplex(swappedResult)

		/*
		 * Apply scaling to the result vector.
		 */
		for i, elem := range result {
			result[i] = scalingFac * elem
		}

		return result

	/*
	 * In-place mode - avoids copies of data elements, faster.
	 */
	case MODE_INPLACE:
		swapComplexInPlace(vec)
		inplaceTransform(vec)
		swapComplexInPlace(vec)

		/*
		 * Apply scaling to the vector.
		 */
		for i, elem := range vec {
			vec[i] = scalingFac * elem
		}

		return vec

	/*
	 * This should never happen.
	 */
	default:
		return nil
	}

}

/*
 * Performs a (forward) Fourier transform of a real-valued vector.
 */
func RealFourier(in []float64, out []complex128, scaling int) error {
	nIn := len(in)
	nOut := len(out)

	/*
	 * Verify that input and output sequences are of equal size.
	 */
	if nIn != nOut {
		return fmt.Errorf("%s", "Input and output sequences must be of equal length.")
	} else {
		m := nIn % 2

		/*
		 * Check if the number of elements in the vector is odd or even.
		 */
		if m != 0 {

			/*
			 * If the number of elements is odd, there may only be a single
			 * element.
			 */
			if nIn == 1 {
				elem := in[0]
				out[0] = complex(elem, 0.0)
				return nil
			} else {
				return fmt.Errorf("%s", "The number of elements in the vector must be even or one.")
			}

		} else {
			nHalf := nIn / 2

			/*
			 * Iterate over the lower half of the output sequence and put
			 * even elements into the real part, odd elements into the
			 * imaginary part of a complex sequence of half the length.
			 */
			for i := 0; i < nHalf; i++ {
				idxEven := i << 1
				idxOdd := idxEven + 1
				even := in[idxEven]
				odd := in[idxOdd]
				out[i] = complex(even, odd)
			}

			lower := out[0:nHalf]
			upper := out[nHalf:nOut]
			Fourier(lower, scaling, MODE_INPLACE)
			copy(upper, lower)
			j := complex(0.0, 1.0)
			coeffs := fourierCoefficients(nIn)

			/*
			 * Iterate over the upper half of the output sequence to perform
			 * an additional butterfly pass and store the result in the lower
			 * half.
			 */
			for i := 0; i < nHalf; i++ {
				idxLow := nHalf + i
				idxHigh := nOut - i

				/*
				 * out[idxHigh] = upper[nHalf - i], but we need to handle
				 * i == 0 specially to stay within the slice bounds.
				 */
				if i == 0 {
					idxHigh = nHalf
				}

				low := out[idxLow]
				high := out[idxHigh]
				highConj := cmplx.Conj(high)
				coeff := j * coeffs[i]
				out[i] = 0.5 * ((low + highConj) - (coeff * (low - highConj)))
			}

			/*
			 * Calculate the remaining parts of the output sequence.
			 */
			for i := 1; i < nHalf; i++ {
				elem := out[i]
				idx := nOut - i
				out[idx] = cmplx.Conj(elem)
			}

			centerElem := out[nHalf]
			centerElemConj := cmplx.Conj(centerElem)
			out[nHalf] = 0.5 * ((centerElem + centerElemConj) + (j * (centerElem - centerElemConj)))

			/*
			 * If we need to apply orthonormal scaling, multiply by inverse
			 * square root of two, to compensate for the larger size of the
			 * transform.
			 */
			if scaling == SCALING_ORTHONORMAL {
				invSqrt2 := complex(1.0/math.Sqrt2, 0.0)

				/*
				 * Multiply each element in the output vector by a square
				 * root of two.
				 */
				for i, elem := range out {
					out[i] = invSqrt2 * elem
				}

			}

			return nil
		}

	}

}

/*
 * Performs an inverse Fourier transform resulting in a real-valued vector.
 *
 * This function will destroy the contents of the input vector in the process.
 */
func RealInverseFourier(in []complex128, out []float64, scaling int) error {
	nIn := len(in)
	nOut := len(out)

	/*
	 * Verify that input and output sequences are of equal size.
	 */
	if nIn != nOut {
		return fmt.Errorf("%s", "Input and output sequences must be of equal length.")
	} else {
		m := nIn % 2

		/*
		 * Check if the number of elements in the vector is odd or even.
		 */
		if m != 0 {

			/*
			 * If the number of elements is odd, there may only be a single
			 * element.
			 */
			if nIn == 1 {
				elem := in[0]
				out[0] = real(elem)
				return nil
			} else {
				return fmt.Errorf("%s", "The number of elements in the vector must be even or one.")
			}

		} else {
			nHalf := nIn / 2

			/*
			 * Ensure that the input array is conjugate symmetric and store
			 * the relevant data in its lower half.
			 */
			for i := 1; i < nHalf; i++ {
				lowValue := in[i]
				idx := nIn - i
				highValue := in[idx]
				highValueConj := cmplx.Conj(highValue)
				average := 0.5 * (lowValue + highValueConj)
				in[i] = average
			}

			/* BEGIN MAGIC */
			dc := in[0]
			dcReal := real(dc)
			nyquist := in[nHalf]
			nyquistReal := real(nyquist)
			/* END MAGIC */

			lower := in[0:nHalf]
			upper := in[nHalf:nIn]
			copy(upper, lower)
			coeffs := fourierCoefficients(nIn)
			j := complex(0.0, 1.0)

			/*
			 * Calculate an inverse butterfly pass on the upper half and
			 * store the results in the lower half of the spectrum.
			 */
			for i := 0; i < nHalf; i++ {
				idxLow := nHalf + i
				idxHigh := nOut - i

				/*
				 * in[idxHigh] = upper[nHalf - i], but we need to handle
				 * i == 0 specially to stay within the slice bounds.
				 */
				if i == 0 {
					idxHigh = nHalf
				}

				low := in[idxLow]
				high := in[idxHigh]
				highConj := cmplx.Conj(high)
				even := low + highConj
				coeff := coeffs[i]
				coeffConj := cmplx.Conj(coeff)
				odd := (low - highConj) * coeffConj
				in[i] = 0.5 * (even + (j * odd))
			}

			/* BEGIN MAGIC */
			firstNewReal := 0.5 * (dcReal + nyquistReal)
			firstNewImag := 0.5 * (dcReal - nyquistReal)
			lower[0] = complex(firstNewReal, firstNewImag)
			/* END MAGIC */

			ZeroComplex(upper)
			InverseFourier(lower, scaling, MODE_INPLACE)

			/*
			 * Extract the real components from the lower half of the
			 * spectrum.
			 */
			for i := 0; i < nHalf; i++ {
				value := in[i]
				idx := i << 1
				idxInc := idx + 1
				out[idx] = real(value)
				out[idxInc] = imag(value)
			}

			/*
			 * If we need to apply orthonormal scaling, multiply by inverse
			 * square root of two, to compensate for the larger size of the
			 * transform.
			 */
			if scaling == SCALING_ORTHONORMAL {

				/*
				 * Multiply each element in the output vector by a square
				 * root of two.
				 */
				for i, elem := range out {
					out[i] = math.Sqrt2 * elem
				}

			}

			return nil
		}

	}

}

/*
 * Shift negative frequencies to lower indices than the DC component or invert the
 * shifting process.
 */
func Shift(vec []complex128, inverse bool) {
	n := len(vec)
	nNegative := n >> 1
	nPositive := nNegative
	isOdd := (n & 1) != 0

	/*
	 * If the number of frequency bins is odd, there is one more bin for positive
	 * frequencies than there are bins for negative frequencies.
	 */
	if isOdd {
		nPositive++
	}

	ptrA := 0
	ptrB := 0

	/*
	 * During the forward operation, the second pointer is offset from the first
	 * pointer by the number of positive coefficients.
	 *
	 * During the inverse operation, the second pointer is offset from the first
	 * pointer by the number of negative coefficients.
	 */
	if inverse {
		ptrB = nNegative
	} else {
		ptrB = nPositive
	}

	/*
	 * Do this until the second pointer reaches the end of the slice.
	 */
	for ptrB < n {
		swapComplexElements(vec, ptrA, ptrB)
		ptrA++
		ptrB++
	}

	/*
	 * If the number of frequency bins is odd, we have to perform further post-
	 * processing. We have to rotate the entire "right half" of the vector,
	 * including the central element, by one position to the left (during the
	 * forward transform) or to the right (during the inverse transform).
	 */
	if isOdd {

		/*
		 * During the forward transform, we have to rotate to the left.
		 * During the inverse transform, we have to rotate to the right.
		 */
		if inverse {
			ptrB = n - 1
			ptrA = ptrB - 1

			/*
			 * Do this until the first pointer reaches the positive elements.
			 */
			for ptrA >= nPositive {
				swapComplexElements(vec, ptrA, ptrB)
				ptrA--
				ptrB--
			}

		} else {
			ptrB = ptrA + 1

			/*
			 * Do this until the second pointer reaches the end of the slice.
			 */
			for ptrB < n {
				swapComplexElements(vec, ptrA, ptrB)
				ptrA++
				ptrB++
			}

		}

	}

}
