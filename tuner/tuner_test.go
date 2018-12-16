package tuner

import (
	"github.com/andrepxx/go-dsp-guitar/wave"
	"io/ioutil"
	"math"
	"testing"
)

/*
 * Perform a unit test on the tuner.
 */
func TestTuner(t *testing.T) {
	tn := Create()

	/*
	 * Paths to test wave files.
	 */
	wavePaths := []string{
		"samples/D2.wav",
		"samples/A2.wav",
		"samples/D3.wav",
		"samples/G3.wav",
		"samples/H3.wav",
		"samples/E4.wav",
	}

	/*
	 * Notes contained in files.
	 */
	notes := []string{
		"D2",
		"A2",
		"D3",
		"G3",
		"H3",
		"E4",
	}

	/*
	 * Import each wave file into a buffer.
	 */
	for i, path := range wavePaths {
		currentNote := notes[i]
		buf, err := ioutil.ReadFile(path)

		/*
		 * Check if file was successfully read.
		 */
		if err != nil {
			t.Errorf("Failed to read wave file from '%s'.", path)
		} else {
			file, err := wave.FromBuffer(buf)

			/*
			 * Check if file was successfully parsed.
			 */
			if err != nil {
				t.Errorf("Failed to parse wave file from '%s'.", path)
			} else {
				sampleRate := file.SampleRate()
				numChannels := file.ChannelCount()

				/*
				 * Check if file has a single channel.
				 */
				if numChannels != 1 {
					t.Errorf("Wave file '%s' has %d channels, expected %d.", path, numChannels, 1)
				} else {
					c, err := file.Channel(0)

					/*
					 * Check if channel could be obtained.
					 */
					if err != nil {
						msg := err.Error()
						t.Errorf("Failed to obtain channel %d from wave file '%s': %s", 1, path, msg)
					} else {
						samples := c.Floats()
						tn.Process(samples, sampleRate)
						res, err := tn.Analyze()

						/*
						 * Check if analysis could be performed.
						 */
						if err != nil {
							msg := err.Error()
							t.Errorf("Failed to analyze wave file '%s': %s", path, msg)
						} else {
							note := res.Note()

							/*
							 * Check if note was determined correctly.
							 */
							if note != currentNote {
								t.Errorf("Tuner failed to determine correct note. Expected '%s', got '%s'.", currentNote, note)
							}

							cents := res.Cents()

							/*
							 * Check if deviation is large.
							 */
							if cents < -5 || cents > 5 {
								t.Errorf("Tuner exhibits large deviation for note '%s'.", currentNote)
							}

							freq := res.Frequency()
							freqInfinite := math.IsInf(freq, 0)
							freqNaN := math.IsNaN(freq)

							/*
							 * Check if frequency is infinite or not a number.
							 */
							if freqInfinite || freqNaN {
								t.Errorf("Tuner reported invalid frequency ('%e') for note '%s'.", freq, currentNote)
							}

						}

					}

				}

			}

		}

	}

}
