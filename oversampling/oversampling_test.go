package oversampling

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
			if diffAbs > 0.0000001 {
				c = false
			}

			diffs[i] = diff
		}

		return c, diffs
	}

}

/*
 * Perform a unit test for two-times oversampling.
 */
func TestTwoTimesOversampling(t *testing.T) {

	/*
	 * Input vectors.
	 */
	in := [][]float64{
		[]float64{0.33155385, 0.60331504, 0.66959906, 0.53356756, 0.4714466, 0.33672882, 0.74765334, 0.80330669},
		[]float64{0.58184733, 0.57457918, 0.28569742, 0.37671357, 0.94136783, 0.65594314, 0.75142587, 0.00428814},
		[]float64{0.75088971, 0.94295395, 0.11823066, 0.85944875, 0.16499603, 0.6089857, 0.60759831, 0.21663312},
		[]float64{0.49558136, 0.65389307, 0.34630074, 0.42389504, 0.34363533, 0.09927411, 0.84365355, 0.41326194},
	}

	oversampledBuffer := make([]float64, 16)
	decimatedBuffer := make([]float64, 8)

	/*
	 * Expected output vectors for two-times oversampling.
	 */
	oversampledExpected := [][]float64{
		[]float64{0.0, 0.0, 0.0, 0.00806242, 0.0, -0.03012038, 0.0, 0.13633848, 0.33155385, 0.49084752, 0.60331504, 0.66842969, 0.66959906, 0.60249345, 0.53356756, 0.50787718},
		[]float64{0.47144660, 0.35404209, 0.33672882, 0.51413626, 0.74765334, 0.84421197, 0.80330669, 0.67858137, 0.58184733, 0.58324543, 0.57457918, 0.43591416, 0.28569742, 0.22800032, 0.37671357, 0.70633113},
		[]float64{0.94136783, 0.82569433, 0.65594314, 0.75524448, 0.75142587, 0.31518428, 0.00428814, 0.24901664, 0.75088971, 1.0523536, 0.94295395, 0.43169085, 0.11823066, 0.47774761, 0.85944875, 0.56224943},
		[]float64{0.16499603, 0.28047679, 0.60898570, 0.72098862, 0.60759831, 0.37176423, 0.21663312, 0.28578320, 0.49558136, 0.64783007, 0.65389307, 0.49745244, 0.34630074, 0.34792703, 0.42389504, 0.44282359},
	}

	/*
	 * Expected output vectors for decimation.
	 */
	decimatedExpected := [][]float64{
		[]float64{0.0, 0.0, -0.00000258, -0.00002004, -0.0000052, -0.00090183, -0.00609235, -0.01285808},
		[]float64{-0.01284382, -0.01074852, -0.00876596, -0.00760304, -0.01380340, -0.01627766, -0.01222176, -0.00884787},
		[]float64{-0.00826708, -0.00634901, -0.01670871, -0.01936035, -0.00059703, -0.02170501, 0.00982247, 0.26768820},
		[]float64{0.58773299, 0.59973714, 0.51166079, 0.41872892, 0.31351970, 0.69258067, 0.75473302, 0.52451811},
	}

	osd := CreateOversamplerDecimator(2)

	/*
	 * Feed input vectors in sequentially.
	 */
	for i, currentIn := range in {
		expectedUp := oversampledExpected[i]
		err := osd.Oversample(currentIn, oversampledBuffer)

		/*
		 * Ensure that there are no errors.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Oversampling vector number %d by factor %d failed: %s", i, 2, msg)
		}

		ok, diff := areSlicesClose(oversampledBuffer, expectedUp)

		/*
		 * Verify components of oversampled vector.
		 */
		if !ok {
			t.Errorf("Oversampling vector number %d by factor %d: Result is incorrect. Expected %v, got %v, difference: %v", i, 2, expectedUp, oversampledBuffer, diff)
		}

		expectedDown := decimatedExpected[i]
		err = osd.Decimate(oversampledBuffer, decimatedBuffer)

		/*
		 * Ensure that there are no errors.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Decimating vector number %d by factor %d failed: %s", i, 2, msg)
		}

		okDown, diffDown := areSlicesClose(decimatedBuffer, expectedDown)

		/*
		 * Verify components of decimated vector.
		 */
		if !okDown {
			t.Errorf("Decimating vector number %d: Result is incorrect. Expected %v, got %v, difference: %v", i, expectedDown, decimatedBuffer, diffDown)
		}

	}

}
