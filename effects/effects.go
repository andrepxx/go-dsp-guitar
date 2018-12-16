package effects

import (
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/filter"
	"math"
	"sync"
)

/*
 * Parameter types.
 */
const (
	PARAMETER_TYPE_INVALID = iota
	PARAMETER_TYPE_DISCRETE
	PARAMETER_TYPE_NUMERIC
)

/*
 * Effect unit types.
 */
const (
	UNIT_SIGNALGENERATOR = iota
	UNIT_NOISEGATE
	UNIT_BANDPASS
	UNIT_AUTOWAH
	UNIT_OCTAVER
	UNIT_EXCESS
	UNIT_FUZZ
	UNIT_OVERDRIVE
	UNIT_DISTORTION
	UNIT_TONESTACK
	UNIT_CHORUS
	UNIT_FLANGER
	UNIT_PHASER
	UNIT_TREMOLO
	UNIT_RINGMODULATOR
	UNIT_DELAY
	UNIT_POWERAMP
)

/*
 * Mathematical constants.
 */
const (
	MATH_DEGREE_TO_RADIANS = math.Pi / 180.0
	MATH_PI_THOUSANDTH     = 0.001 * math.Pi
	MATH_TWO_OVER_PI       = 2.0 / math.Pi
	MATH_TWO_PI            = 2.0 * math.Pi
	MATH_TWO_PI_FIFTH      = 0.4 * math.Pi
	MATH_TWO_PI_HUNDREDTH  = 0.02 * math.Pi
)

/*
 * Other constants.
 */
const (
	NUM_FILTERS = 8
	STRING_NONE = "- NONE -"
)

/*
 * Data structure representing the result of a filter compilation.
 */
type compilationResult struct {
	id     uint64
	result filter.Filter
}

/*
 * Data structure representing a parameter for an effects unit.
 */
type Parameter struct {
	Name               string
	Type               int32
	Minimum            int32
	Maximum            int32
	NumericValue       int32
	DiscreteValueIndex int
	DiscreteValues     []string
}

/*
 * Interface type for an effects unit.
 */
type Unit interface {
	Parameters() []Parameter
	Process(in []float64, out []float64, sampleRate uint32)
	Type() int
	SetDiscreteValue(name string, value string) error
	GetDiscreteValue(name string) (string, error)
	SetNumericValue(name string, value int32) error
	GetNumericValue(name string) (int32, error)
}

/*
 * Data structure representing a generic effects unit.
 */
type unitStruct struct {
	unitType int
	mutex    sync.RWMutex
	params   []Parameter
}

/*
 * Returns the parameters of an effects unit.
 */
func (this *unitStruct) parameters() []Parameter {
	n := len(this.params)
	params := make([]Parameter, n)
	copy(params, this.params)

	/*
	 * Copy the discrete value slices.
	 */
	for i, param := range params {
		values := param.DiscreteValues
		k := len(values)
		valuesCopy := make([]string, k)
		copy(valuesCopy, values)
		params[i].DiscreteValues = valuesCopy
	}

	return params
}

/*
 * Returns the parameters of an effects unit.
 */
func (this *unitStruct) Parameters() []Parameter {
	this.mutex.RLock()
	params := this.parameters()
	this.mutex.RUnlock()
	return params
}

/*
 * Returns the type of this effects unit.
 */
func (this *unitStruct) Type() int {
	return this.unitType
}

/*
 * Sets a discrete parameter value for an effects unit.
 */
func (this *unitStruct) setDiscreteValue(name string, value string) error {
	idx := int(-1)

	/*
	 * Iterate over all parameters.
	 */
	for i, param := range this.params {

		/*
		 * If we got the right one, store its index.
		 */
		if param.Name == name {
			idx = i
		}

	}

	/*
	 * Check if parameter was found.
	 */
	if idx == -1 {
		return fmt.Errorf("Failed to set discrete value: Could not find parameter with name '%s'.", name)
	} else {
		param := this.params[idx]

		/*
		 * Check if parameter is discrete.
		 */
		if param.Type != PARAMETER_TYPE_DISCRETE {
			return fmt.Errorf("Failed to set discrete value: Parameter '%s' is not discrete.", name)
		} else {
			values := param.DiscreteValues
			valIdx := int(-1)

			/*
			 * Iterate over all values.
			 */
			for i, currentValue := range values {

				/*
				 * If we got the right one, store its index.
				 */
				if currentValue == value {
					valIdx = i
				}

			}

			/*
			 * Check if discrete value was found.
			 */
			if valIdx == -1 {
				return fmt.Errorf("Failed to set discrete value: Value '%s' is not valid for parameter '%s'.", value, name)
			} else {
				this.params[idx].DiscreteValueIndex = valIdx
				return nil
			}

		}

	}

}

/*
 * Sets a discrete parameter value for an effects unit.
 */
func (this *unitStruct) SetDiscreteValue(name string, value string) error {
	this.mutex.Lock()
	err := this.setDiscreteValue(name, value)
	this.mutex.Unlock()
	return err
}

/*
 * Gets a discrete parameter value from an effects unit.
 */
func (this *unitStruct) getDiscreteValue(name string) (string, error) {
	idx := int(-1)

	/*
	 * Iterate over all parameters.
	 */
	for i, param := range this.params {

		/*
		 * If we got the right one, store its index.
		 */
		if param.Name == name {
			idx = i
		}

	}

	/*
	 * Check if parameter was found.
	 */
	if idx == -1 {
		return "", fmt.Errorf("Failed to get discrete value: Could not find parameter with name '%s'.", name)
	} else {
		param := this.params[idx]

		/*
		 * Check if parameter is discrete.
		 */
		if param.Type != PARAMETER_TYPE_DISCRETE {
			return "", fmt.Errorf("Failed to get discrete value: Parameter '%s' is not discrete.", name)
		} else {
			valIdx := param.DiscreteValueIndex
			value := param.DiscreteValues[valIdx]
			return value, nil
		}

	}

}

/*
 * Gets a discrete parameter value from an effects unit.
 */
func (this *unitStruct) GetDiscreteValue(name string) (string, error) {
	this.mutex.RLock()
	val, err := this.getDiscreteValue(name)
	this.mutex.RUnlock()
	return val, err
}

/*
 * Sets a numeric parameter value for an effects unit.
 */
func (this *unitStruct) setNumericValue(name string, value int32) error {
	idx := int(-1)

	/*
	 * Iterate over all parameters.
	 */
	for i, param := range this.params {

		/*
		 * If we got the right one, store its index.
		 */
		if param.Name == name {
			idx = i
		}

	}

	/*
	 * Check if parameter was found.
	 */
	if idx == -1 {
		return fmt.Errorf("Failed to set numeric value: Could not find parameter with name '%s'.", name)
	} else {
		param := this.params[idx]

		/*
		 * Check if parameter is numeric.
		 */
		if param.Type != PARAMETER_TYPE_NUMERIC {
			return fmt.Errorf("Failed to set numeric value: Parameter '%s' is not numeric.", name)
		} else {
			min := param.Minimum
			max := param.Maximum

			/*
			 * Check if value is out of range.
			 */
			if (value < min) || (value > max) {
				return fmt.Errorf("Failed to set numeric value: Parameter '%s' must be between '%d' and '%d' - got '%d'.", name, min, max, value)
			} else {
				this.params[idx].NumericValue = value
				return nil
			}

		}

	}

}

/*
 * Sets a numeric parameter value for an effects unit.
 */
func (this *unitStruct) SetNumericValue(name string, value int32) error {
	this.mutex.Lock()
	err := this.setNumericValue(name, value)
	this.mutex.Unlock()
	return err
}

/*
 * Gets a numeric parameter value from an effects unit.
 */
func (this *unitStruct) getNumericValue(name string) (int32, error) {
	idx := int(-1)

	/*
	 * Iterate over all parameters.
	 */
	for i, param := range this.params {

		/*
		 * If we got the right one, store its index.
		 */
		if param.Name == name {
			idx = i
		}

	}

	/*
	 * Check if parameter was found.
	 */
	if idx == -1 {
		return 0, fmt.Errorf("Failed to get numeric value: Could not find parameter with name '%s'.", name)
	} else {
		param := this.params[idx]

		/*
		 * Check if parameter is numeric.
		 */
		if param.Type != PARAMETER_TYPE_NUMERIC {
			return 0, fmt.Errorf("Failed to get numeric value: Parameter '%s' is not numeric.", name)
		} else {
			val := param.NumericValue
			return val, nil
		}

	}

}

/*
 * Gets a numeric parameter value from an effects unit.
 */
func (this *unitStruct) GetNumericValue(name string) (int32, error) {
	this.mutex.RLock()
	val, err := this.getNumericValue(name)
	this.mutex.RUnlock()
	return val, err
}

/*
 * Turn gain (or attenuation) in decibels into a (linear) factor.
 */
func decibelsToFactor(decibels int32) float64 {
	decibelsFloat := float64(decibels)
	exp := 0.05 * decibelsFloat
	result := math.Pow(10.0, exp)
	return result
}

/*
 * Turn a linear factor into a gain (or attenuation) value in decibels.
 */
func factorToDecibels(factor float64) float64 {
	result := 20.0 * math.Log10(factor)
	return result
}

/*
 * Returns the sign of an integer.
 */
func signInt(number int32) float64 {

	/*
	 * Check, whether number is negative, positive or zero.
	 */
	if number < 0 {
		return -1.0
	} else if number > 0 {
		return 1.0
	} else {
		return 0
	}

}

/*
 * Returns the sign of a floating-point number.
 */
func signFloat(number float64) float64 {

	/*
	 * Check, whether number is negative, positive or zero.
	 */
	if number < 0.0 {
		return -1.0
	} else if number > 0.0 {
		return 1.0
	} else {
		return 0.0
	}

}

/*
 * Create a new effects unit.
 */
func CreateUnit(unitType int) Unit {

	/*
	 * Lookup, which effect unit to create.
	 */
	switch unitType {
	case UNIT_SIGNALGENERATOR:
		u := createSignalGenerator()
		return u
	case UNIT_NOISEGATE:
		u := createNoiseGate()
		return u
	case UNIT_BANDPASS:
		u := createBandpass()
		return u
	case UNIT_AUTOWAH:
		u := createAutoWah()
		return u
	case UNIT_OCTAVER:
		u := createOctaver()
		return u
	case UNIT_EXCESS:
		u := createExcess()
		return u
	case UNIT_FUZZ:
		u := createFuzz()
		return u
	case UNIT_OVERDRIVE:
		u := createOverdrive()
		return u
	case UNIT_DISTORTION:
		u := createDistortion()
		return u
	case UNIT_TONESTACK:
		u := createToneStack()
		return u
	case UNIT_CHORUS:
		u := createChorus()
		return u
	case UNIT_FLANGER:
		u := createFlanger()
		return u
	case UNIT_PHASER:
		u := createPhaser()
		return u
	case UNIT_TREMOLO:
		u := createTremolo()
		return u
	case UNIT_RINGMODULATOR:
		u := createRingModulator()
		return u
	case UNIT_DELAY:
		u := createDelay()
		return u
	case UNIT_POWERAMP:
		u := createPowerAmp()
		return u
	default:
		return nil
	}

}

/*
 * Returns a list of supported parameter types.
 */
func ParameterTypes() []string {

	/*
	 * List of all supported parameter types.
	 */
	paramTypes := []string{
		"invalid",
		"discrete",
		"numeric",
	}

	return paramTypes
}

/*
 * Returns a list of supported unit types.
 */
func UnitTypes() []string {

	/*
	 * List of all supported unit types.
	 */
	unitTypes := []string{
		"signal_generator",
		"noise_gate",
		"bandpass",
		"auto_wah",
		"octaver",
		"excess",
		"fuzz",
		"overdrive",
		"distortion",
		"tone_stack",
		"chorus",
		"flanger",
		"phaser",
		"tremolo",
		"ring_modulator",
		"delay",
		"power_amp",
	}

	return unitTypes
}
