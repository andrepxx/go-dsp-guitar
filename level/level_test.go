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
 * Perform a unit test on the level meter.
 */
func TestMeter(t *testing.T) {
	sampleRate := uint32(DEFAULT_SAMPLE_RATE)
	sampleRateFloat := float64(sampleRate)
	buf := make([]float64, sampleRate)

	/*
	 * Generate data series.
	 */
	for i := range buf {
		iFloat := float64(i)
		t := iFloat / sampleRateFloat
		arg := TWO_PI * t
		elem := math.Sin(arg)
		buf[i] = elem
	}

	m := CreateMeter()
	m.Process(buf, sampleRate)
	res := m.Analyze()
	level := res.Level()
	peak := res.Peak()
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
