package metronome

import (
	"sync"
)

/*
 * Global constants.
 */
const (
	DEFAULT_SAMPLE_RATE = 96000
	OUTPUT_COUNT        = 1
)

/*
 * Data structure representing a metronome.
 */
type metronomeStruct struct {
	sampleCounter    uint32
	tickCounter      uint32
	mutex            sync.RWMutex
	beatsPerPeriod   uint32
	bpmSpeed         uint32
	coefficientsTick []float64
	coefficientsTock []float64
	nameTick         string
	nameTock         string
	sampleRate       uint32
}

/*
 * Interface type representing a metronome.
 */
type Metronome interface {
	BeatsPerPeriod() uint32
	Process(outputBuffer []float64)
	SampleRate() uint32
	SetBeatsPerPeriod(count uint32)
	SetSampleRate(rate uint32)
	SetSpeed(speed uint32)
	SetTick(name string, coefficients []float64)
	SetTock(name string, coefficients []float64)
	Tick() (string, []float64)
	Tock() (string, []float64)
	Speed() uint32
}

/*
 * Returns the number of beats per period for this metronome.
 */
func (this *metronomeStruct) BeatsPerPeriod() uint32 {
	this.mutex.RLock()
	bpm := this.beatsPerPeriod
	this.mutex.RUnlock()
	return bpm
}

/*
 * Generates the metronome signal and writes it into a buffer.
 */
func (this *metronomeStruct) Process(outputBuffer []float64) {
	this.mutex.RLock()
	tickBuf := this.coefficientsTick
	tockBuf := this.coefficientsTock
	bpm := this.bpmSpeed
	beatsPerPeriod := this.beatsPerPeriod
	this.mutex.RUnlock()

	/*
	 * Prevent division by zero.
	 */
	if beatsPerPeriod == 0 {
		beatsPerPeriod = 1
	}

	sampleCounter := this.sampleCounter
	tickCounter := this.tickCounter
	sampleRate := this.sampleRate
	tickSize := len(tickBuf)
	unsignedTickSize := uint32(tickSize)
	tockSize := len(tockBuf)
	unsignedTockSize := uint32(tockSize)
	samplesPerBeat := (60 * sampleRate) / bpm

	/*
	 * Generate the output samples.
	 */
	for i, _ := range outputBuffer {
		sample := float64(0.0)

		/*
		 * Decide whether a tick or a tock should be produced.
		 */
		if tickCounter == 0 {

			/*
			 * Check if buffer is allocated and part of the tick must be output.
			 */
			if (tickBuf != nil) && (sampleCounter < unsignedTickSize) {
				sample = tickBuf[sampleCounter]
			}

		} else {

			/*
			 * Check if buffer is allocated and part of the tock must be output.
			 */
			if (tockBuf != nil) && (sampleCounter < unsignedTockSize) {
				sample = tockBuf[sampleCounter]
			}

		}

		outputBuffer[i] = sample
		sampleCounter++

		/*
		 * Reset sample counter on every beat.
		 */
		if sampleCounter > samplesPerBeat {
			sampleCounter = 0
			tickCounter = (tickCounter + 1) % beatsPerPeriod
		}

	}

	this.sampleCounter = sampleCounter
	this.tickCounter = tickCounter
}

/*
 * Return the sample rate the metronome is supposed to operate at.
 */
func (this *metronomeStruct) SampleRate() uint32 {
	this.mutex.RLock()
	rate := this.sampleRate
	this.mutex.RUnlock()
	return rate
}

/*
 * Sets the number of beats per period.
 */
func (this *metronomeStruct) SetBeatsPerPeriod(count uint32) {
	this.mutex.Lock()
	this.beatsPerPeriod = count
	this.mutex.Unlock()
}

/*
 * Sets the sample rate. Note that the coefficients will also need to be
 * updated on a sample rate change.
 */
func (this *metronomeStruct) SetSampleRate(rate uint32) {
	this.mutex.Lock()
	this.sampleRate = rate
	this.mutex.Unlock()
}

/*
 * Sets the speed in beats per minute.
 */
func (this *metronomeStruct) SetSpeed(speed uint32) {
	this.mutex.Lock()
	this.bpmSpeed = speed
	this.mutex.Unlock()
}

/*
 * Set the name and the coefficients for the 'tick' signal.
 */
func (this *metronomeStruct) SetTick(name string, coefficients []float64) {
	this.mutex.Lock()
	this.nameTick = name

	/*
	 * Check if coefficients were passed into the function.
	 */
	if coefficients == nil {
		this.coefficientsTick = nil
	} else {
		size := len(coefficients)

		/*
		 * If size of the coefficient buffer does not match, allocate
		 * a new one.
		 */
		if size != len(this.coefficientsTick) {
			this.coefficientsTick = make([]float64, size)
		}

		copy(this.coefficientsTick, coefficients)
	}

	this.mutex.Unlock()
}

/*
 * Set the name and the coefficients for the 'tock' signal.
 */
func (this *metronomeStruct) SetTock(name string, coefficients []float64) {
	this.mutex.Lock()
	this.nameTock = name

	/*
	 * Check if coefficients were passed into the function.
	 */
	if coefficients == nil {
		this.coefficientsTock = nil
	} else {
		size := len(coefficients)

		/*
		 * If size of the coefficient buffer does not match, allocate
		 * a new one.
		 */
		if size != len(this.coefficientsTock) {
			this.coefficientsTock = make([]float64, size)
		}

		copy(this.coefficientsTock, coefficients)
	}

	this.mutex.Unlock()
}

/*
 * Returns the name and the coefficients of the metronome 'tick' sound.
 */
func (this *metronomeStruct) Tick() (string, []float64) {
	this.mutex.RLock()
	coeffs := this.coefficientsTick
	size := len(coeffs)
	coeffsCopy := make([]float64, size)
	copy(coeffsCopy, coeffs)
	this.mutex.RUnlock()
	return this.nameTick, coeffsCopy
}

/*
 * Returns the name and the coefficients of the metronome 'tock' sound.
 */
func (this *metronomeStruct) Tock() (string, []float64) {
	this.mutex.RLock()
	coeffs := this.coefficientsTock
	size := len(coeffs)
	coeffsCopy := make([]float64, size)
	copy(coeffsCopy, coeffs)
	this.mutex.RUnlock()
	return this.nameTock, coeffsCopy
}

/*
 * Returns the metronome speed in beats per minute.
 */
func (this *metronomeStruct) Speed() uint32 {
	this.mutex.RLock()
	bpm := this.bpmSpeed
	this.mutex.RUnlock()
	return bpm
}

/*
 * Creates a new metronome.
 */
func Create() Metronome {

	/*
	 * Create a new metronome struct.
	 */
	m := metronomeStruct{
		beatsPerPeriod:   4,
		bpmSpeed:         120,
		coefficientsTick: nil,
		coefficientsTock: nil,
		sampleCounter:    0,
		sampleRate:       DEFAULT_SAMPLE_RATE,
		tickCounter:      0,
	}

	return &m
}
