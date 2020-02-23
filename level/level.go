package level

import (
	"fmt"
	"math"
	"sync"
)

/*
 * Global constants.
 */
const (
	PEAK_HOLD_TIME_SECONDS = 2
	TIME_CONSTANT          = 1.7 // DIN IEC 60268-18
	MIN_LEVEL              = -200.0
	OUTPUT_COUNT           = 1
)

/*
 * Data structure representing the result of a level analysis.
 */
type resultStruct struct {
	level int32
	peak  int32
}

/*
 * The result of a level analysis.
 */
type Result interface {
	Level() int32
	Peak() int32
}

/*
 * Data structure representing a level meter for a single channel.
 */
type channelMeterStruct struct {
	channelName   string
	mutex         sync.RWMutex
	enabled       bool
	currentValue  float64
	peakValue     float64
	sampleCounter uint64
}

/*
 * Data structure representing level meters for multiple channels.
 */
type meterStruct struct {
	channelMeters []*channelMeterStruct
	mutex         sync.RWMutex
	enabled       bool
}

/*
 * Interface type representing a level meter for muliple channels.
 */
type Meter interface {
	Analyze(channelId uint32) (Result, error)
	ChannelCount() uint32
	ChannelName(channelId uint32) (string, error)
	Enabled() bool
	Process(inputBuffers [][]float64, sampleRate uint32) error
	SetEnabled(value bool)
}

/*
 * Turn a linear factor into a gain (or attenuation) value in decibels.
 */
func factorToDecibels(factor float64) float64 {
	result := 20.0 * math.Log10(factor)
	return result
}

/*
 * Returns the current signal level.
 */
func (this *resultStruct) Level() int32 {
	value := this.level
	return value
}

/*
 * Returns the current peak level.
 */
func (this *resultStruct) Peak() int32 {
	value := this.peak
	return value
}

/*
 * Perform analysis of signal level of a single channel.
 */
func (this *channelMeterStruct) analyze() Result {
	this.mutex.RLock()
	currentValue := this.currentValue
	peakValue := this.peakValue
	this.mutex.RUnlock()
	currentLevel := factorToDecibels(currentValue)
	currentLevelNaN := math.IsNaN(currentLevel)

	/*
	 * Ensure that the minimum level is not exceeded.
	 */
	if currentLevelNaN || currentLevel < MIN_LEVEL {
		currentLevel = MIN_LEVEL
	}

	currentLevelRounded := math.Round(currentLevel)
	currentLevelInt := int32(currentLevelRounded)
	peakLevel := factorToDecibels(peakValue)
	peakLevelNaN := math.IsNaN(peakLevel)

	/*
	 * Ensure that the minimum level is not exceeded.
	 */
	if peakLevelNaN || peakLevel < MIN_LEVEL {
		peakLevel = MIN_LEVEL
	}

	peakLevelRounded := math.Round(peakLevel)
	peakLevelInt := int32(peakLevelRounded)

	/*
	 * Create result structure.
	 */
	result := resultStruct{
		level: currentLevelInt,
		peak:  peakLevelInt,
	}

	return &result
}

/*
 * Returns the name of the channel measured by this channel meter.
 */
func (this *channelMeterStruct) name() string {
	name := this.channelName
	return name
}

/*
 * Feed the signal from an input buffer through a single-channel level meter.
 */
func (this *channelMeterStruct) process(buffer []float64, sampleRate uint32) {
	this.mutex.RLock()
	enabled := this.enabled
	this.mutex.RUnlock()

	/*
	 * Only perform processing if this channel is enabled.
	 */
	if enabled {
		this.mutex.RLock()
		currentValue := this.currentValue
		peakValue := this.peakValue
		sampleCounter := this.sampleCounter
		this.mutex.RUnlock()
		sampleRateFloat := float64(sampleRate)
		holdTimeSamples := uint64(PEAK_HOLD_TIME_SECONDS * sampleRateFloat)
		decayExp := -1.0 / (TIME_CONSTANT * sampleRateFloat)
		decayFactor := math.Pow(10.0, decayExp)

		/*
		 * Process each sample.
		 */
		for _, sample := range buffer {
			currentValue *= decayFactor

			/*
			 * If we're above the hold time, let the peak indicator decay,
			 * otherwise increment sample counter.
			 */
			if sampleCounter > holdTimeSamples {
				peakValue *= decayFactor
			} else {
				sampleCounter++
			}

			sampleAbs := math.Abs(sample)

			/*
			 * If we got a sample with larger amplitude, update current value.
			 */
			if sampleAbs > currentValue {
				currentValue = sampleAbs
			}

			/*
			 * If we got a sample with larger or equal amplitude, update peak value.
			 */
			if sampleAbs >= peakValue {
				peakValue = sampleAbs
				sampleCounter = 0
			}

		}

		this.mutex.Lock()
		this.currentValue = currentValue
		this.peakValue = peakValue
		this.sampleCounter = sampleCounter
		this.mutex.Unlock()
	}

}

/*
 * Enables or disables level measurements for this channel.
 */
func (this *channelMeterStruct) setEnabled(value bool) {
	this.mutex.Lock()
	enabled := this.enabled

	/*
	 * Check if status of meter must be changed.
	 */
	if value != enabled {

		/*
		 * If level meter should be disabled, clear state.
		 */
		if !value {
			this.currentValue = 0.0
			this.peakValue = 0.0
			this.sampleCounter = 0
		}

		this.enabled = value
	}

	this.mutex.Unlock()
}

/*
 * Analyze the level of a certain channel.
 */
func (this *meterStruct) Analyze(channelId uint32) (Result, error) {
	channelMeters := this.channelMeters
	numMeters := len(channelMeters)
	numMeters32 := uint32(numMeters)

	/*
	 * Check if channel number is within range.
	 */
	if channelId >= numMeters32 {
		return nil, fmt.Errorf("Requested analysis for channel %d, but level meter only has %d channels.", channelId, numMeters)
	} else {
		channelMeter := channelMeters[channelId]
		res := channelMeter.analyze()
		return res, nil
	}

}

/*
 * Returns the number of channels this meter is able to process.
 */
func (this *meterStruct) ChannelCount() uint32 {
	channelMeters := this.channelMeters
	numChannels := len(channelMeters)
	numChannels32 := uint32(numChannels)
	return numChannels32
}

/*
 * Returns the name of the channel with the provided id.
 */
func (this *meterStruct) ChannelName(channelId uint32) (string, error) {
	channelMeters := this.channelMeters
	numMeters := len(channelMeters)
	numMeters32 := uint32(numMeters)

	/*
	 * Check if channel number is within range.
	 */
	if channelId >= numMeters32 {
		return "", fmt.Errorf("Requested name of channel %d, but level meter only has %d channels.", channelId, numMeters)
	} else {
		channelMeter := channelMeters[channelId]
		name := channelMeter.name()
		return name, nil
	}

}

/*
 * Returns whether this level meter is enabled.
 */
func (this *meterStruct) Enabled() bool {
	this.mutex.RLock()
	enabled := this.enabled
	this.mutex.RUnlock()
	return enabled
}

/*
 * Process input buffers for multiple channels.
 */
func (this *meterStruct) Process(buffers [][]float64, sampleRate uint32) error {
	channelMeters := this.channelMeters
	numChannels := len(channelMeters)
	numBuffers := len(buffers)

	/*
	 * Make sure that the correct number of buffers is provided.
	 */
	if numChannels != numBuffers {
		return fmt.Errorf("Number of input buffers (%d) does not match number of channels (%d) for this level meter.", numBuffers, numChannels)
	} else {

		/*
		 * Feed input from each channel to the corresponding level meter.
		 */
		for i, buffer := range buffers {
			channelMeter := channelMeters[i]
			channelMeter.process(buffer, sampleRate)
		}

		return nil
	}

}

/*
 * Enables or disables this level meter.
 */
func (this *meterStruct) SetEnabled(value bool) {
	this.mutex.Lock()
	enabled := this.enabled

	/*
	 * Check if value must be changed.
	 */
	if value != enabled {
		channelMeters := this.channelMeters

		/*
		 * Enable or disable each channel meter.
		 */
		for _, channelMeter := range channelMeters {
			channelMeter.setEnabled(value)
		}

		this.enabled = value
	}

	this.mutex.Unlock()
}

/*
 * Creates a new level meter for a certain number of channels.
 */
func CreateMeter(numChannels uint32, names []string) (Meter, error) {
	numNames := len(names)
	numNames32 := uint32(numNames)

	/*
	 * Check if number of channel names matches number of channels.
	 */
	if numChannels != numNames32 {
		return nil, fmt.Errorf("Failed to create channel meter. Requested channel meter for %d channels, but provided %d channel names.", numChannels, numNames)
	} else {
		channelMeters := make([]*channelMeterStruct, numChannels)

		/*
		 * Create the channel meters.
		 */
		for i := range channelMeters {
			name := names[i]

			/*
			 * Create a new channel meter.
			 */
			channelMeter := &channelMeterStruct{
				channelName:   name,
				enabled:       false,
				currentValue:  0.0,
				peakValue:     0.0,
				sampleCounter: 0,
			}

			channelMeters[i] = channelMeter
		}

		/*
		 * Create a new level meter.
		 */
		meter := meterStruct{
			channelMeters: channelMeters,
			enabled:       false,
		}

		return &meter, nil
	}

}
