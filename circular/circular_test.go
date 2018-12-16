package circular

import (
	"testing"
)

/*
 * Compare two slices to check whether their contents are equal.
 */
func areSlicesEqual(a []float64, b []float64) bool {
	eq := true

	/*
	 * Check whether the two slices are of the same size.
	 */
	if len(a) != len(b) {
		eq = false
	} else {

		/*
		 * Iterate over the arrays to compare values.
		 */
		for i, elem := range a {

			/*
			 * Check if we found a mismatch.
			 */
			if b[i] != elem {
				eq = false
			}

		}

	}

	return eq
}

/*
 * Perform a unit test on the circular buffer.
 */
func TestBuffer(t *testing.T) {
	bufSize := 5
	buf := CreateBuffer(bufSize)

	/*
	 * Make sure the buffer is non-nil.
	 */
	if buf == nil {
		t.Errorf("Buffer returned from circular.CreateBuffer(5) == nil.")
	} else {

		/*
		 * Describe the elements to be enqueued.
		 */
		in := [][]float64{
			[]float64{1.0},
			[]float64{2.0, 3.0},
			[]float64{4.0, 5.0, 6.0},
			[]float64{7.0, 8.0},
			[]float64{9.0, 10.0},
			[]float64{11.0, 12.0, 13.0, 14.0, 15.0},
			[]float64{16.0, 17.0, 18.0, 19.0, 20.0, 21.0},
			[]float64{31.0, 32.0, 33.0, 34.0},
			[]float64{35.0, 36.0, 37.0, 38.0},
			[]float64{39.0, 40.0, 41.0, 42.0},
			[]float64{43.0},
			[]float64{44.0},
		}

		out := make([][]float64, 7)

		/*
		 * Initialize inner slices.
		 */
		for i, _ := range out {
			out[i] = make([]float64, bufSize)
		}

		errs := make([]error, 7)
		buf.Enqueue(in[0]...)
		errs[0] = buf.Retrieve(out[0])
		buf.Enqueue(in[1]...)
		errs[1] = buf.Retrieve(out[1])
		buf.Enqueue(in[2]...)
		errs[2] = buf.Retrieve(out[2])
		buf.Enqueue(in[3]...)
		buf.Enqueue(in[4]...)
		errs[3] = buf.Retrieve(out[3])
		buf.Enqueue(in[5]...)
		errs[4] = buf.Retrieve(out[4])
		buf.Enqueue(in[6]...)
		errs[5] = buf.Retrieve(out[5])
		buf.Enqueue(in[7]...)
		buf.Enqueue(in[8]...)
		buf.Enqueue(in[9]...)
		buf.Enqueue(in[10][0])
		buf.Enqueue(in[11][0])
		errs[6] = buf.Retrieve(out[6])

		/*
		 * Describe the elements we expect to be retrieved.
		 */
		expected := [][]float64{
			[]float64{0.0, 0.0, 0.0, 0.0, 1.0},
			[]float64{0.0, 0.0, 1.0, 2.0, 3.0},
			[]float64{2.0, 3.0, 4.0, 5.0, 6.0},
			[]float64{6.0, 7.0, 8.0, 9.0, 10.0},
			[]float64{11.0, 12.0, 13.0, 14.0, 15.0},
			[]float64{17.0, 18.0, 19.0, 20.0, 21.0},
			[]float64{40.0, 41.0, 42.0, 43.0, 44.0},
		}

		/*
		 * Compare the results.
		 */
		for i, elem := range out {
			ex := expected[i]
			pass := areSlicesEqual(elem, ex)

			/*
			 * If slices differ, report that.
			 */
			if !pass {
				t.Errorf("Array number %d retrieved from circular buffer does not match expectations. Expected: %v Got: %v", i, ex, elem)
			}

		}

		/*
		 * Check for errors.
		 */
		for i, err := range errs {

			/*
			 * Check if an error occured.
			 */
			if err != nil {
				msg := err.Error()
				t.Errorf("Error while retrieving array %d: %s", i, msg)
			}

		}

		tooSmall := make([]float64, bufSize-1)
		err := buf.Retrieve(tooSmall)

		/*
		 * Verify that an error was returned when retrieving contents into a too small buffer.
		 */
		if err == nil {
			t.Errorf("Did not return an error when retrieving into too small buffer.")
		}

		l := buf.Length()

		/*
		 * Check if returned length is correct.
		 */
		if l != bufSize {
			t.Errorf("Circular buffer did not return correct length.")
		}

	}

}
