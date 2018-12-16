package level

import (
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
 * Data structure representing a level meter.
 */
type meterStruct struct {
	mutex         sync.RWMutex
	currentValue  float64
	peakValue     float64
	sampleCounter uint64
}

/*
 * Interface type representing a level meter.
 */
type Meter interface {
	Process(inputBuffer []float64, sampleRate uint32)
	Analyze() Result
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
 * Feed the signal from an input buffer through the level meter.
 */
func (this *meterStruct) Process(inputBuffer []float64, sampleRate uint32) {
	this.mutex.Lock()
	currentValue := this.currentValue
	peakValue := this.peakValue
	sampleCounter := this.sampleCounter
	sampleRateFloat := float64(sampleRate)
	holdTimeSamples := uint64(PEAK_HOLD_TIME_SECONDS * sampleRateFloat)
	decayExp := -1.0 / (TIME_CONSTANT * sampleRateFloat)
	decayFactor := math.Pow(10.0, decayExp)

	/*
	 * Process each sample.
	 */
	for _, sample := range inputBuffer {
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

	this.currentValue = currentValue
	this.peakValue = peakValue
	this.sampleCounter = sampleCounter
	this.mutex.Unlock()
}

/*
 * Perform analysis of signal level.
 */
func (this *meterStruct) Analyze() Result {
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
 * Creates a new level meter.
 */
func CreateMeter() Meter {

	/*
	 * Create a new meter struct.
	 */
	m := meterStruct{
		currentValue:  0.0,
		peakValue:     0.0,
		sampleCounter: 0,
	}

	return &m
}
