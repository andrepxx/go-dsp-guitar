package signal

import (
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/effects"
	"github.com/andrepxx/go-dsp-guitar/filter"
	"sync"
)

/*
 * Data structure representing a slot in a signal chain.
 */
type slotStruct struct {
	unit   effects.Unit
	bypass bool
}

/*
 * Interface type for a signal chain.
 */
type Chain interface {
	AppendUnit(unitType int) (int, error)
	RemoveUnit(id int) error
	MoveUp(id int) error
	MoveDown(id int) error
	UnitType(id int) (int, error)
	SetBypass(id int, bypass bool) error
	GetBypass(id int) (bool, error)
	SetDiscreteValue(id int, name string, value string) error
	GetDiscreteValue(id int, name string) (string, error)
	SetNumericValue(id int, name string, value int32) error
	GetNumericValue(id int, name string) (int32, error)
	Parameters(id int) ([]effects.Parameter, error)
	Length() int
	Process(in []float64, out []float64, sampleRate uint32)
}

/*
 * Data structure representing a signal chain.
 */
type chainStruct struct {
	bufferIn  []float64
	bufferOut []float64
	responses filter.ImpulseResponses
	mutex     sync.RWMutex
	slots     []slotStruct
}

/*
 * Appends a new effects unit to the end of the signal chain.
 */
func (this *chainStruct) AppendUnit(unitType int) (int, error) {
	unit := effects.CreateUnit(unitType)

	/*
	 * Check whether unit was successfully created.
	 */
	if unit == nil {
		return -1, fmt.Errorf("%s", "Failed to create effects unit.")
	} else {

		/*
		 * If unit is a power amp, prepare it.
		 */
		if unitType == effects.UNIT_POWERAMP {
			effects.PreparePowerAmp(unit, this.responses)
		}

		/*
		 * Create new slot in the signal chain.
		 */
		slot := slotStruct{
			unit:   unit,
			bypass: true,
		}

		this.mutex.Lock()
		nPre := len(this.slots)
		this.slots = append(this.slots, slot)
		this.mutex.Unlock()
		return nPre, nil
	}

}

/*
 * Removes an effects unit from the signal chain.
 */
func (this *chainStruct) RemoveUnit(id int) error {
	this.mutex.Lock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.Unlock()
		return fmt.Errorf("Cannot remove unit %d.", id)
	} else {
		idInc := id + 1
		this.slots = append(this.slots[:id], this.slots[idInc:]...)
		this.mutex.Unlock()
		return nil
	}

}

/*
 * Moves an effects unit up (towards the front of) the signal chain.
 */
func (this *chainStruct) MoveUp(id int) error {
	this.mutex.Lock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id <= 0 || id >= n {
		this.mutex.Unlock()
		return fmt.Errorf("Cannot move unit %d up.", id)
	} else {
		idDec := id - 1
		this.slots[id], this.slots[idDec] = this.slots[idDec], this.slots[id]
		this.mutex.Unlock()
		return nil
	}

}

/*
 * Moves an effects unit down (towards the end of) the signal chain.
 */
func (this *chainStruct) MoveDown(id int) error {
	this.mutex.Lock()
	n := len(this.slots)
	nDec := n - 1

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= nDec {
		this.mutex.Unlock()
		return fmt.Errorf("Cannot move unit %d down.", id)
	} else {
		idInc := id + 1
		this.slots[id], this.slots[idInc] = this.slots[idInc], this.slots[id]
		this.mutex.Unlock()
		return nil
	}

}

/*
 * Returns the type of an effects unit.
 */
func (this *chainStruct) UnitType(id int) (int, error) {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.RUnlock()
		return -1, fmt.Errorf("Cannot get unit type: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		unitType := unit.Type()
		this.mutex.RUnlock()
		return unitType, nil
	}
}

/*
 * Enables or disables bypass of an effects unit inside the signal chain.
 */
func (this *chainStruct) SetBypass(id int, bypass bool) error {
	this.mutex.Lock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.Unlock()
		action := "disable"

		/*
		 * Check whether bypass should be enabled.
		 */
		if bypass {
			action = "enable"
		}

		return fmt.Errorf("Cannot %s bypass: No unit %d.", action, id)
	} else {
		this.slots[id].bypass = bypass
		this.mutex.Unlock()
		return nil
	}

}

/*
 * Retrieves whether an effects unit inside the signal chain is in bypass mode or not.
 */
func (this *chainStruct) GetBypass(id int) (bool, error) {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.RUnlock()
		return false, fmt.Errorf("Cannot get bypass value: No unit %d.", id)
	} else {
		bypass := this.slots[id].bypass
		this.mutex.RUnlock()
		return bypass, nil
	}

}

/*
 * Sets a discrete value for an effects unit inside the signal chain.
 */
func (this *chainStruct) SetDiscreteValue(id int, name string, value string) error {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.RUnlock()
		return fmt.Errorf("Cannot set discrete value: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		this.mutex.RUnlock()
		err := unit.SetDiscreteValue(name, value)
		return err
	}

}

/*
 * Retrieves a discrete value from an effects unit inside the signal chain.
 */
func (this *chainStruct) GetDiscreteValue(id int, name string) (string, error) {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		this.mutex.RUnlock()
		return "", fmt.Errorf("Cannot get discrete value: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		this.mutex.RUnlock()
		value, err := unit.GetDiscreteValue(name)
		return value, err
	}

}

/*
 * Sets a numeric value for an effects unit inside the signal chain.
 */
func (this *chainStruct) SetNumericValue(id int, name string, value int32) error {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		return fmt.Errorf("Cannot set numeric value: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		this.mutex.RUnlock()
		err := unit.SetNumericValue(name, value)
		return err
	}

}

/*
 * Retrieves a numeric value from an effects unit inside the signal chain.
 */
func (this *chainStruct) GetNumericValue(id int, name string) (int32, error) {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		return 0, fmt.Errorf("Cannot get numeric value: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		this.mutex.RUnlock()
		value, err := unit.GetNumericValue(name)
		return value, err
	}

}

/*
 * Returns the parameters of an effects unit inside a signal chain.
 */
func (this *chainStruct) Parameters(id int) ([]effects.Parameter, error) {
	this.mutex.RLock()
	n := len(this.slots)

	/*
	 * Check if index is out of range.
	 */
	if id < 0 || id >= n {
		return nil, fmt.Errorf("Cannot get parameters: No unit %d.", id)
	} else {
		unit := this.slots[id].unit
		this.mutex.RUnlock()
		params := unit.Parameters()
		return params, nil
	}

}

/*
 * Returns the number of units inside this signal chain.
 */
func (this *chainStruct) Length() int {
	this.mutex.RLock()
	n := len(this.slots)
	this.mutex.RUnlock()
	return n
}

/*
 * Passes a signal through the signal chain.
 */
func (this *chainStruct) Process(in []float64, out []float64, sampleRate uint32) {

	/*
	 * Verify that input and output buffers are the same size.
	 */
	if len(in) == len(out) {
		n := len(in)

		/*
		 * If size of input buffer does not match, reallocate it.
		 */
		if len(this.bufferIn) != n {
			this.bufferIn = make([]float64, n)
		}

		/*
		 * If size of output buffer does not match, reallocate it.
		 */
		if len(this.bufferOut) != n {
			this.bufferOut = make([]float64, n)
		}

		copy(this.bufferIn, in)
		this.mutex.RLock()

		/*
		 * Iterate over the slots.
		 */
		for _, slot := range this.slots {

			/*
			 * Verify that slot is not in bypass mode.
			 */
			if !slot.bypass {
				slot.unit.Process(this.bufferIn, this.bufferOut, sampleRate)
				this.bufferIn, this.bufferOut = this.bufferOut, this.bufferIn
			}

		}

		this.mutex.RUnlock()
		copy(out, this.bufferIn)
	}

}

/*
 * Creates a new signal chain.
 */
func CreateChain(responses filter.ImpulseResponses) Chain {
	slots := make([]slotStruct, 0)

	/*
	 * The new signal chain.
	 */
	chain := chainStruct{
		responses: responses,
		slots:     slots,
	}

	return &chain
}
