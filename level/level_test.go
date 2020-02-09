package level

import (
	"math"
	"testing"
)

const (
	DEFAULT_SAMPLE_RATE = 96000
	TESTING_FREQUENCY   = 440
	TWO_PI              = 2.0 * math.Pi
)

/*
 * Perform a unit test on the level meters.
 */
func TestMeters(t *testing.T) {
	sampleRate := uint32(DEFAULT_SAMPLE_RATE)
	sampleRateFloat := float64(sampleRate)
	bufA := make([]float64, sampleRate)
	bufB := make([]float64, sampleRate)

	/*
	 * Generate data series.
	 */
	for i := uint32(0); i < sampleRate; i++ {
		iFloat := float64(i)
		t := iFloat / sampleRateFloat
		arg := TWO_PI * t
		elemA := math.Sin(arg)
		elemB := 0.5 * elemA
		bufA[i] = elemA
		bufB[i] = elemB
	}

	/*
	 * Channel buffers.
	 */
	bufs := [][]float64{
		bufA,
		bufB,
	}

	/*
	 * Channel names.
	 */
	names := []string{
		"channel_a",
		"channel_b",
	}

	m, err := CreateMeter(2, names)

	/*
	 * Check if level meter was sucessfully created.
	 */
	if err != nil {
		msg := err.Error()
		t.Errorf("Creating %d channel level meter failed: %s", 2, msg)
	} else {
		nameA, err := m.ChannelName(0)

		/*
		 * Verify availability of channel name.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Obtaining name of channel %d returned error: %s", 0, msg)
		} else {

			/*
			 * Verify channel name.
			 */
			if nameA != "channel_a" {
				t.Errorf("Name of channel %d incorrect. Expected: '%s' Got: '%s'", 0, "channel_a", nameA)
			}

		}

		nameB, err := m.ChannelName(1)

		/*
		 * Verify availability of channel name.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Obtaining name of channel %d returned error: %s", 0, msg)
		} else {

			/*
			 * Verify channel name.
			 */
			if nameB != "channel_b" {
				t.Errorf("Name of channel %d incorrect. Expected: '%s' Got: '%s'", 0, "channel_b", nameB)
			}

		}

		m.Process(bufs, sampleRate)
		resA, err := m.Analyze(0)

		/*
		 * Check if level analysis returned error.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Level meter analysis for channel %d returned error: %s", 0, msg)
		} else {
			level := resA.Level()
			peak := resA.Peak()
			expectedLevel := int32(-3)
			expectedPeak := int32(0)

			/*
			 * Check if the current level matches our expectations.
			 */
			if level != expectedLevel {
				t.Errorf("Current level does not match! Expected %d, got %d.\n", expectedLevel, level)
			}

			/*
			 * Check if the peak level matches our expectations.
			 */
			if peak != expectedPeak {
				t.Errorf("Peak level does not match! Expected %d, got %d.\n", expectedPeak, peak)
			}

		}

		resB, err := m.Analyze(1)

		/*
		 * Check if level analysis returned error.
		 */
		if err != nil {
			msg := err.Error()
			t.Errorf("Level meter analysis for channel %d returned error: %s", 1, msg)
		} else {
			level := resB.Level()
			peak := resB.Peak()
			expectedLevel := int32(-9)
			expectedPeak := int32(-6)

			/*
			 * Check if the current level matches our expectations.
			 */
			if level != expectedLevel {
				t.Errorf("Current level does not match! Expected %d, got %d.\n", expectedLevel, level)
			}

			/*
			 * Check if the peak level matches our expectations.
			 */
			if peak != expectedPeak {
				t.Errorf("Peak level does not match! Expected %d, got %d.\n", expectedPeak, peak)
			}

		}

	}

}
