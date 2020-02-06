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

/*
 * Perform a unit test for four-times oversampling.
 */
func TestFourTimesOversampling(t *testing.T) {

	/*
	 * Input vectors.
	 */
	in := [][]float64{
		[]float64{0.33155385, 0.60331504, 0.66959906, 0.53356756, 0.4714466, 0.33672882, 0.74765334, 0.80330669},
		[]float64{0.58184733, 0.57457918, 0.28569742, 0.37671357, 0.94136783, 0.65594314, 0.75142587, 0.00428814},
		[]float64{0.75088971, 0.94295395, 0.11823066, 0.85944875, 0.16499603, 0.6089857, 0.60759831, 0.21663312},
		[]float64{0.49558136, 0.65389307, 0.34630074, 0.42389504, 0.34363533, 0.09927411, 0.84365355, 0.41326194},
	}

	oversampledBuffer := make([]float64, 32)
	decimatedBuffer := make([]float64, 8)

	/*
	 * Expected output vectors for four-times oversampling.
	 */
	oversampledExpected := [][]float64{
		[]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.00243889, 0.00806242, 0.00995361, 0.0, -0.01803854, -0.03012038, -0.02594172, 0.0, 0.05360846, 0.13633848, 0.23504417, 0.33155385, 0.41664661, 0.49084752, 0.55362129, 0.60331504, 0.64115246, 0.66842969, 0.67977884, 0.66959906, 0.64046007, 0.60249345, 0.56483648, 0.53356756, 0.51410523, 0.50787718, 0.50053240},
		[]float64{0.47144660, 0.41503086, 0.35404209, 0.32062025, 0.33672882, 0.40491903, 0.51413626, 0.63913871, 0.74765334, 0.81670109, 0.84421197, 0.83758462, 0.80330669, 0.74612363, 0.67858137, 0.61895299, 0.58184733, 0.57224104, 0.58324543, 0.59301412, 0.57457918, 0.51679874, 0.43591416, 0.35420674, 0.28569742, 0.23820529, 0.22800032, 0.27243303, 0.37671357, 0.52999710, 0.70633113, 0.85992621},
		[]float64{0.94136783, 0.92272366, 0.82569433, 0.71503105, 0.65594314, 0.67832264, 0.75524448, 0.80697582, 0.75142587, 0.56711556, 0.31518428, 0.09783875, 0.00428814, 0.06349449, 0.24901664, 0.50164387, 0.75088971, 0.94341463, 1.05235357, 1.05750390, 0.94295395, 0.71454474, 0.43169085, 0.19989503, 0.11823066, 0.22799315, 0.47774761, 0.73487034, 0.85944875, 0.78533184, 0.56224943, 0.31531734},
		[]float64{0.16499603, 0.16115649, 0.28047679, 0.45499623, 0.60898570, 0.69904300, 0.72098862, 0.68657535, 0.60759831, 0.49458598, 0.37176423, 0.27069681, 0.21663312, 0.22248833, 0.28578320, 0.38643535, 0.49558136, 0.58687434, 0.64783007, 0.67240578, 0.65389307, 0.59002212, 0.49745244, 0.40689649, 0.34630074, 0.32819221, 0.34792703, 0.38750026, 0.42389504, 0.44323496, 0.44282359, 0.41386116},
	}

	/*
	 * Expected output vectors for decimation.
	 */
	decimatedExpected := [][]float64{
		[]float64{0.0, 0.0, -0.00000092, -0.00001574, -0.00000539, -0.00038558, -0.00420028, -0.01144640},
		[]float64{-0.01331591, -0.01100871, -0.00943473, -0.00712265, -0.01183699, -0.01636720, -0.01334079, -0.00927324},
		[]float64{-0.00843790, -0.00633489, -0.01255269, -0.02204667, -0.00326029, -0.01363751, -0.01488744, 0.182251478},
		[]float64{0.52760944, 0.62458510, 0.52375668, 0.45929084, 0.30219552, 0.57700540, 0.81172630, 0.54158997},
	}

	osd := CreateOversamplerDecimator(4)

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
			t.Errorf("Oversampling vector number %d by factor %d failed: %s", i, 4, msg)
		}

		ok, diff := areSlicesClose(oversampledBuffer, expectedUp)

		/*
		 * Verify components of oversampled vector.
		 */
		if !ok {
			t.Errorf("Oversampling vector number %d by factor %d: Result is incorrect. Expected %v, got %v, difference: %v", i, 4, expectedUp, oversampledBuffer, diff)
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
