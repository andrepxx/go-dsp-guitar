package random

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
 * Perform a unit test on the random number generator.
 */
func TestRNG(t *testing.T) {

	/*
	 * The seeds with which the PRNG is tested.
	 */
	seeds := []uint64{
		0,
		1,
		1337,
		0xffffffffffffffff,
	}

	/*
	 * The values we expect from the PRNG output.
	 */
	expectedOutputs := [][]float64{
		[]float64{0.000649588648834814, 0.9176364163101058, 0.7152417425208183, 0.06796094967793762, 0.2196807053123421, 0.17361246531234353, 0.9047031462236337, 0.34577150023148534},
		[]float64{0.5091992369938635, 0.11157217073400708, 0.1934726533419198, 0.6948832037811011, 0.9020005109738564, 0.92258087864386, 0.8168201472766885, 0.29620888670553347},
		[]float64{0.931529109768131, 0.20974058258323053, 0.10996983489950173, 0.26301429538336984, 0.48126045007376045, 0.5443806234229176, 0.405133608640296, 0.08055724676750343},
		[]float64{0.4921312462465197, 0.24985181377255528, 0.25943212002462906, 0.27563922365721244, 0.6684298498261998, 0.3004807977010317, 0.18076460965048952, 0.11079298109821321},
	}

	outputBuffer := make([]float64, 8)

	/*
	 * Initialize the PRNG with each of the seeds and obtain its output.
	 */
	for i, seed := range seeds {
		rng := CreatePRNG(seed)

		/*
		 * Fill the output buffer with values from the PRNG.
		 */
		for i, _ := range outputBuffer {
			outputBuffer[i] = rng.NextFloat()
		}

		expectedOutput := expectedOutputs[i]
		valid, diff := areSlicesClose(outputBuffer, expectedOutput)

		/*
		 * Check if we got the expected result.
		 */
		if !valid {
			t.Errorf("PRNG test number %d: Result is incorrect. Seeded with %d, expected %v, got %v, difference: %v", i, seed, expectedOutput, outputBuffer, diff)
		}

		/*
		 * Obtain 10k more values from the PRNG and verify that they are all within the unit interval.
		 */
		for j := 0; j < 10000; j++ {
			value := rng.NextFloat()

			/*
			 * Check if the interval is exceeded.
			 */
			if value < 0.0 || value > 1.0 {
				t.Errorf("PRNG test number %d, seeded with %d exceeded unit interval [0; 1] at the %d-th sample. Output: %f", i, seed, j, value)
			}

		}

	}

}
