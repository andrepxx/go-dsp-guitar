package wave

import (
	"fmt"
	"math"
	"testing"
)

/*
 * Compare two real-valued slices to check whether their components are close.
 */
func areSlicesClose(a []float64, b []float64, err float64) (bool, []float64) {

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
			if diffAbs > err {
				c = false
			}

			diffs[i] = diff
		}

		return c, diffs
	}

}

/*
 * Compare two byte slices to check whether they are equal.
 */
func areSlicesEqual(a []byte, b []byte) bool {

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
 * Convert buffer to hex string.
 */
func bufferToHex(buf []byte) string {
	s := "["

	/*
	 * Serialize all bytes into their hexadecimal representation.
	 */
	for i, b := range buf {

		/*
		 * Prepend comma if this is not the first element.
		 */
		if i > 0 {
			s += ", "
		}

		s += fmt.Sprintf("0x%02x", b)
	}

	s += "]"
	return s
}

/*
 * Test creating an 8-bit mono PCM wave file.
 */
func TestExportPCM8Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0x38, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0x77, 0x01, 0x00,
		0x01, 0x00, 0x08, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x14, 0x00, 0x00, 0x00, 0x57, 0x87, 0x01, 0x20,
		0xd5, 0x5c, 0xea, 0x34, 0x06, 0x40, 0x6d, 0xff,
		0x9a, 0x8d, 0x94, 0xa6, 0x80, 0x76, 0xd6, 0x8c,
	}

	w, err := CreateEmpty(96000, AUDIO_PCM, 8, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading an 8-bit mono PCM wave file.
 */
func TestImportPCM8Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0x38, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0x77, 0x01, 0x00,
		0x01, 0x00, 0x08, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x14, 0x00, 0x00, 0x00, 0x57, 0x87, 0x01, 0x20,
		0xd5, 0x5c, 0xea, 0x34, 0x06, 0x40, 0x6d, 0xff,
		0x9a, 0x8d, 0x94, 0xa6, 0x80, 0x76, 0xd6, 0x8c,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_PCM {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_PCM, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 8 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 8, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 0.078125)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}

/*
 * Test creating a 16-bit mono PCM wave file.
 */
func TestExportPCM16Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0x4c, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xee, 0x02, 0x00,
		0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x28, 0x00, 0x00, 0x00, 0xfc, 0xd5, 0xe5, 0x07,
		0x01, 0x80, 0x6a, 0x9e, 0x3d, 0x56, 0x34, 0xdb,
		0x68, 0x6b, 0x04, 0xb3, 0xb9, 0x84, 0x49, 0xbf,
		0x5d, 0xec, 0xff, 0x7f, 0xf0, 0x1a, 0x74, 0x0d,
		0x1a, 0x15, 0x20, 0x27, 0x00, 0x00, 0xbc, 0xf5,
		0xa9, 0x57, 0x54, 0x0c,
	}

	w, err := CreateEmpty(96000, AUDIO_PCM, 16, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading an 16-bit mono PCM wave file.
 */
func TestImportPCM16Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0x4c, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xee, 0x02, 0x00,
		0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x28, 0x00, 0x00, 0x00, 0xfc, 0xd5, 0xe5, 0x07,
		0x01, 0x80, 0x6a, 0x9e, 0x3d, 0x56, 0x34, 0xdb,
		0x68, 0x6b, 0x04, 0xb3, 0xb9, 0x84, 0x49, 0xbf,
		0x5d, 0xec, 0xff, 0x7f, 0xf0, 0x1a, 0x74, 0x0d,
		0x1a, 0x15, 0x20, 0x27, 0x00, 0x00, 0xbc, 0xf5,
		0xa9, 0x57, 0x54, 0x0c,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_PCM {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_PCM, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 16 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 16, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 3.0518e-5)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}

/*
 * Test creating a 24-bit mono PCM wave file.
 */
func TestExportPCM24Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0x60, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0x65, 0x04, 0x00,
		0x03, 0x00, 0x18, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x3c, 0x00, 0x00, 0x00, 0x9d, 0xfb, 0xd5, 0xac,
		0xe5, 0x07, 0x01, 0x00, 0x80, 0xf7, 0x68, 0x9e,
		0x84, 0x3d, 0x56, 0x3c, 0x33, 0xdb, 0xe3, 0x68,
		0x6b, 0x9e, 0x03, 0xb3, 0x4e, 0xb8, 0x84, 0x7d,
		0x48, 0xbf, 0x49, 0x5c, 0xec, 0xff, 0xff, 0x7f,
		0x4f, 0xf0, 0x1a, 0x86, 0x74, 0x0d, 0xb6, 0x1a,
		0x15, 0xdf, 0x20, 0x27, 0x00, 0x00, 0x00, 0x51,
		0xbb, 0xf5, 0x79, 0xa9, 0x57, 0x37, 0x54, 0x0c,
	}

	w, err := CreateEmpty(96000, AUDIO_PCM, 24, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading a 24-bit mono PCM wave file.
 */
func TestImportPCM24Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0x60, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0x65, 0x04, 0x00,
		0x03, 0x00, 0x18, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x3c, 0x00, 0x00, 0x00, 0x9d, 0xfb, 0xd5, 0xac,
		0xe5, 0x07, 0x01, 0x00, 0x80, 0xf7, 0x68, 0x9e,
		0x84, 0x3d, 0x56, 0x3c, 0x33, 0xdb, 0xe3, 0x68,
		0x6b, 0x9e, 0x03, 0xb3, 0x4e, 0xb8, 0x84, 0x7d,
		0x48, 0xbf, 0x49, 0x5c, 0xec, 0xff, 0xff, 0x7f,
		0x4f, 0xf0, 0x1a, 0x86, 0x74, 0x0d, 0xb6, 0x1a,
		0x15, 0xdf, 0x20, 0x27, 0x00, 0x00, 0x00, 0x51,
		0xbb, 0xf5, 0x79, 0xa9, 0x57, 0x37, 0x54, 0x0c,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_PCM {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_PCM, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 24 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 24, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 1.1921e-7)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}

/*
 * Test creating a 32-bit mono PCM wave file.
 */
func TestExportPCM32Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0x74, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xdc, 0x05, 0x00,
		0x04, 0x00, 0x20, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x50, 0x00, 0x00, 0x00, 0xaf, 0x9c, 0xfb, 0xd5,
		0x97, 0xac, 0xe5, 0x07, 0x01, 0x00, 0x00, 0x80,
		0xe4, 0xf5, 0x68, 0x9e, 0x46, 0x85, 0x3d, 0x56,
		0x6c, 0x3b, 0x33, 0xdb, 0x6d, 0xe3, 0x68, 0x6b,
		0x19, 0x9d, 0x03, 0xb3, 0xe4, 0x4c, 0xb8, 0x84,
		0xdb, 0x7b, 0x48, 0xbf, 0x96, 0x48, 0x5c, 0xec,
		0xff, 0xff, 0xff, 0x7f, 0x5d, 0x4f, 0xf0, 0x1a,
		0x0e, 0x86, 0x74, 0x0d, 0x10, 0xb7, 0x1a, 0x15,
		0x73, 0xdf, 0x20, 0x27, 0x00, 0x00, 0x00, 0x00,
		0x79, 0x50, 0xbb, 0xf5, 0x0c, 0x7a, 0xa9, 0x57,
		0x8f, 0x37, 0x54, 0x0c,
	}

	w, err := CreateEmpty(96000, AUDIO_PCM, 32, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading a 32-bit mono PCM wave file.
 */
func TestImportPCM32Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0x74, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xdc, 0x05, 0x00,
		0x04, 0x00, 0x20, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x50, 0x00, 0x00, 0x00, 0xaf, 0x9c, 0xfb, 0xd5,
		0x97, 0xac, 0xe5, 0x07, 0x01, 0x00, 0x00, 0x80,
		0xe4, 0xf5, 0x68, 0x9e, 0x46, 0x85, 0x3d, 0x56,
		0x6c, 0x3b, 0x33, 0xdb, 0x6d, 0xe3, 0x68, 0x6b,
		0x19, 0x9d, 0x03, 0xb3, 0xe4, 0x4c, 0xb8, 0x84,
		0xdb, 0x7b, 0x48, 0xbf, 0x96, 0x48, 0x5c, 0xec,
		0xff, 0xff, 0xff, 0x7f, 0x5d, 0x4f, 0xf0, 0x1a,
		0x0e, 0x86, 0x74, 0x0d, 0x10, 0xb7, 0x1a, 0x15,
		0x73, 0xdf, 0x20, 0x27, 0x00, 0x00, 0x00, 0x00,
		0x79, 0x50, 0xbb, 0xf5, 0x0c, 0x7a, 0xa9, 0x57,
		0x8f, 0x37, 0x54, 0x0c,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_PCM {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_PCM, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 32 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 32, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 4.6567e-10)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}

/*
 * Test creating a 32-bit IEEE floating-point wave file.
 */
func TestExportIEEE32Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0x74, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xdc, 0x05, 0x00,
		0x04, 0x00, 0x20, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x50, 0x00, 0x00, 0x00, 0x8d, 0x11, 0xa8, 0xbe,
		0x93, 0xb5, 0x7c, 0x3d, 0x00, 0x00, 0x80, 0xbf,
		0x14, 0x2e, 0x43, 0xbf, 0x0b, 0x7b, 0x2c, 0x3f,
		0x12, 0x33, 0x93, 0xbe, 0xc7, 0xd1, 0x56, 0x3f,
		0xc6, 0xf8, 0x19, 0xbf, 0x66, 0x8f, 0x76, 0xbf,
		0x08, 0x6f, 0x01, 0xbf, 0xbb, 0x1d, 0x1d, 0xbe,
		0x00, 0x00, 0x80, 0x3f, 0x7b, 0x82, 0x57, 0x3e,
		0x61, 0x48, 0xd7, 0x3d, 0xb9, 0xd5, 0x28, 0x3e,
		0x7e, 0x83, 0x9c, 0x3e, 0x00, 0x00, 0x00, 0x00,
		0xf8, 0x4a, 0xa4, 0xbd, 0xf4, 0x52, 0x2f, 0x3f,
		0x79, 0x43, 0xc5, 0x3d,
	}

	w, err := CreateEmpty(96000, AUDIO_IEEE_FLOAT, 32, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading an 32-bit IEEE floating-point wave file.
 */
func TestImportIEEE32Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0x74, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xdc, 0x05, 0x00,
		0x04, 0x00, 0x20, 0x00, 0x64, 0x61, 0x74, 0x61,
		0x50, 0x00, 0x00, 0x00, 0x8d, 0x11, 0xa8, 0xbe,
		0x93, 0xb5, 0x7c, 0x3d, 0x00, 0x00, 0x80, 0xbf,
		0x14, 0x2e, 0x43, 0xbf, 0x0b, 0x7b, 0x2c, 0x3f,
		0x12, 0x33, 0x93, 0xbe, 0xc7, 0xd1, 0x56, 0x3f,
		0xc6, 0xf8, 0x19, 0xbf, 0x66, 0x8f, 0x76, 0xbf,
		0x08, 0x6f, 0x01, 0xbf, 0xbb, 0x1d, 0x1d, 0xbe,
		0x00, 0x00, 0x80, 0x3f, 0x7b, 0x82, 0x57, 0x3e,
		0x61, 0x48, 0xd7, 0x3d, 0xb9, 0xd5, 0x28, 0x3e,
		0x7e, 0x83, 0x9c, 0x3e, 0x00, 0x00, 0x00, 0x00,
		0xf8, 0x4a, 0xa4, 0xbd, 0xf4, 0x52, 0x2f, 0x3f,
		0x79, 0x43, 0xc5, 0x3d,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_IEEE_FLOAT {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_IEEE_FLOAT, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 32 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 32, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 1.1921e-7)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}

/*
 * Test creating a 64-bit IEEE floating-point wave file.
 */
func TestExportIEEE64Mono(t *testing.T) {

	/*
	 * Sample data for testing.
	 */
	samples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	/*
	 * Expected output buffer.
	 */
	expectedOutput := []byte{
		0x52, 0x49, 0x46, 0x46, 0xc4, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xb8, 0x0b, 0x00,
		0x08, 0x00, 0x40, 0x00, 0x64, 0x61, 0x74, 0x61,
		0xa0, 0x00, 0x00, 0x00, 0xd5, 0x84, 0xc4, 0xa8,
		0x31, 0x02, 0xd5, 0xbf, 0x51, 0x7d, 0x8c, 0x5e,
		0xb2, 0x96, 0xaf, 0x3f, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0xbf, 0x61, 0x93, 0x4e, 0x87,
		0xc2, 0x65, 0xe8, 0xbf, 0xe6, 0x95, 0xa9, 0x51,
		0x61, 0x8f, 0xe5, 0x3f, 0x1b, 0x0d, 0x22, 0x4a,
		0x62, 0x66, 0xd2, 0xbf, 0x07, 0xba, 0x93, 0xdb,
		0x38, 0xda, 0xea, 0x3f, 0x0b, 0x36, 0xe0, 0xb9,
		0x18, 0x3f, 0xe3, 0xbf, 0x93, 0x17, 0x3e, 0xc7,
		0xec, 0xd1, 0xee, 0xbf, 0x57, 0xc0, 0x6f, 0x09,
		0xe1, 0x2d, 0xe0, 0xbf, 0x8a, 0x05, 0x3a, 0x6a,
		0xb7, 0xa3, 0xc3, 0xbf, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0x3f, 0xb4, 0x31, 0xe1, 0x5d,
		0x4f, 0xf0, 0xca, 0x3f, 0xa0, 0x9a, 0x9a, 0x1d,
		0x0c, 0xe9, 0xba, 0x3f, 0x55, 0xf2, 0x77, 0x10,
		0xb7, 0x1a, 0xc5, 0x3f, 0x08, 0x3f, 0xcc, 0xb9,
		0x6f, 0x90, 0xd3, 0x3f, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xbe, 0x36, 0xd9, 0x0e,
		0x5f, 0x89, 0xb4, 0xbf, 0xe2, 0x22, 0x18, 0x83,
		0x5e, 0xea, 0xe5, 0x3f, 0x0f, 0x8c, 0x72, 0x1f,
		0x6f, 0xa8, 0xb8, 0x3f,
	}

	w, err := CreateEmpty(96000, AUDIO_IEEE_FLOAT, 64, 1)

	/*
	 * Check if wave file was successfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to create wave file: %s", msg)
	} else {
		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				c.WriteFloats(samples)
				buf, err := w.Bytes()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if err != nil {
					t.Errorf("%s", "Attempt to obtain byte buffer failed.")
				} else {

					/*
					 * Make sure that buffer is non-nil.
					 */
					if buf == nil {
						t.Errorf("%s", "Byte buffer is nil.")
					} else {
						equal := areSlicesEqual(buf, expectedOutput)

						/*
						 * If buffers are not equal, report failure.
						 */
						if !equal {
							expectedOutputString := bufferToHex(expectedOutput)
							actualOutputString := bufferToHex(buf)
							t.Errorf("Byte buffers are not equal. Expected: %s Got: %s", expectedOutputString, actualOutputString)
						}

					}

				}

			}

		}

	}

}

/*
 * Test reading an 64-bit IEEE floating-point wave file.
 */
func TestImportIEEE64Mono(t *testing.T) {

	/*
	 * Input buffer.
	 */
	buf := []byte{
		0x52, 0x49, 0x46, 0x46, 0xc4, 0x00, 0x00, 0x00,
		0x57, 0x41, 0x56, 0x45, 0x66, 0x6d, 0x74, 0x20,
		0x10, 0x00, 0x00, 0x00, 0x03, 0x00, 0x01, 0x00,
		0x00, 0x77, 0x01, 0x00, 0x00, 0xb8, 0x0b, 0x00,
		0x08, 0x00, 0x40, 0x00, 0x64, 0x61, 0x74, 0x61,
		0xa0, 0x00, 0x00, 0x00, 0xd5, 0x84, 0xc4, 0xa8,
		0x31, 0x02, 0xd5, 0xbf, 0x51, 0x7d, 0x8c, 0x5e,
		0xb2, 0x96, 0xaf, 0x3f, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0xbf, 0x61, 0x93, 0x4e, 0x87,
		0xc2, 0x65, 0xe8, 0xbf, 0xe6, 0x95, 0xa9, 0x51,
		0x61, 0x8f, 0xe5, 0x3f, 0x1b, 0x0d, 0x22, 0x4a,
		0x62, 0x66, 0xd2, 0xbf, 0x07, 0xba, 0x93, 0xdb,
		0x38, 0xda, 0xea, 0x3f, 0x0b, 0x36, 0xe0, 0xb9,
		0x18, 0x3f, 0xe3, 0xbf, 0x93, 0x17, 0x3e, 0xc7,
		0xec, 0xd1, 0xee, 0xbf, 0x57, 0xc0, 0x6f, 0x09,
		0xe1, 0x2d, 0xe0, 0xbf, 0x8a, 0x05, 0x3a, 0x6a,
		0xb7, 0xa3, 0xc3, 0xbf, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xf0, 0x3f, 0xb4, 0x31, 0xe1, 0x5d,
		0x4f, 0xf0, 0xca, 0x3f, 0xa0, 0x9a, 0x9a, 0x1d,
		0x0c, 0xe9, 0xba, 0x3f, 0x55, 0xf2, 0x77, 0x10,
		0xb7, 0x1a, 0xc5, 0x3f, 0x08, 0x3f, 0xcc, 0xb9,
		0x6f, 0x90, 0xd3, 0x3f, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xbe, 0x36, 0xd9, 0x0e,
		0x5f, 0x89, 0xb4, 0xbf, 0xe2, 0x22, 0x18, 0x83,
		0x5e, 0xea, 0xe5, 0x3f, 0x0f, 0x8c, 0x72, 0x1f,
		0x6f, 0xa8, 0xb8, 0x3f,
	}

	/*
	 * Expected sample data.
	 */
	expectedSamples := []float64{
		-0.32825891, 0.0616966, -1.0, -0.76242186,
		0.67375246, -0.28749902, 0.83913844, -0.60145222,
		-0.9631256, -0.50560047, -0.15343373, 1.0,
		0.21045868, 0.10511852, 0.16487778, 0.3056907,
		0.0, -0.08022112, 0.68485952, 0.0963201,
	}

	w, err := FromBuffer(buf)

	/*
	 * Check if wave file was read created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Failed to read wave file: %s", msg)
	} else {
		sampleRate := w.SampleRate()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if sampleRate != 96000 {
			t.Errorf("Attempt to determine sample rate failed. Expected %d, got %d.", 96000, sampleRate)
		}

		numChannels := w.ChannelCount()

		/*
		 * Check if sample rate was determined successfully.
		 */
		if numChannels != 1 {
			t.Errorf("Attempt to determine channel count failed. Expected %d, got %d.", 1, numChannels)
		}

		sampleFormat := w.SampleFormat()

		/*
		 * Check if sample format was determined successfully.
		 */
		if sampleFormat != AUDIO_IEEE_FLOAT {
			t.Errorf("Attempt to determine sample format failed. Expected %d, got %d.", AUDIO_IEEE_FLOAT, sampleFormat)
		}

		bitDepth := w.BitDepth()

		/*
		 * Check if bit depth was determined successfully.
		 */
		if bitDepth != 64 {
			t.Errorf("Attempt to determine bit depth failed. Expected %d, got %d.", 64, bitDepth)
		}

		c, err := w.Channel(1)

		/*
		 * Attempt to obtain non-existing channel must return nil reference.
		 */
		if c != nil {
			t.Errorf("Attempt to obtain non-existant channel did not return nil.")
		}

		/*
		 * Attempt to obtain non-existing channel must return error.
		 */
		if err == nil {
			t.Errorf("%s", "Attempt to obtain non-existant channel did not return error.")
		}

		c, err = w.Channel(0)

		/*
		 * Attempt to obtain existing channel must not return error.
		 */
		if err != nil {
			t.Errorf("%s", "Attempt to obtain existing channel returned error.")
		} else {

			/*
			 * Attempt to obtain existing channel must not return nil reference.
			 */
			if c == nil {
				t.Errorf("%s", "Attempt to obtain existing channel returned nil.")
			} else {
				samples := c.Floats()

				/*
				 * Check if attempt to obtain byte buffer was successful.
				 */
				if samples == nil {
					t.Errorf("%s", "Sample buffer is nil.")
				} else {
					equal, diff := areSlicesClose(samples, expectedSamples, 1.0e-16)

					/*
					 * If buffers are not equal, report failure.
					 */
					if !equal {
						t.Errorf("Sample buffers are not similar. Expected: %v Got: %v Difference: %v", expectedSamples, samples, diff)
					}

				}

			}

		}

	}

}
