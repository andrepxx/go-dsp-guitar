package circular

import (
	"fmt"
	"sync"
)

/*
 * Data structure implementing a circular buffer.
 */
type bufferStruct struct {
	mutex   sync.RWMutex
	values  []float64
	pointer int
}

/*
 * A circular buffer.
 */
type Buffer interface {
	Enqueue(elems ...float64)
	Length() int
	Retrieve(buf []float64) error
}

/*
 * Add elements to the circular buffer, potentially overwriting unread elements.
 *
 * Semantics: First write to buffer, then increment pointer.
 *
 * Pointer points to "oldest" element, or next element to be overwritten.
 */
func (this *bufferStruct) Enqueue(elems ...float64) {
	numElems := len(elems)
	values := this.values
	n := len(values)

	/*
	 * If there are more elements than fit into the buffer, simply copy
	 * the tail of the element array into the buffer, otherwise perform
	 * circular write operation.
	 */
	if numElems >= n {
		idx := numElems - n
		this.mutex.Lock()
		copy(values, elems[idx:numElems])
		this.pointer = 0
		this.mutex.Unlock()
	} else {
		this.mutex.Lock()
		ptr := this.pointer
		ptrInc := ptr + numElems

		/*
		 * Check whether the write operation stays within the array bounds.
		 */
		if ptrInc < n {
			copy(values[ptr:ptrInc], elems)
			this.pointer = ptrInc
		} else {
			head := ptrInc - n
			tail := n - ptr
			copy(values[ptr:n], elems[0:tail])
			copy(values[0:head], elems[tail:numElems])
			this.pointer = head
		}

		this.mutex.Unlock()
	}

}

/*
 * Returns the size of this buffer.
 */
func (this *bufferStruct) Length() int {
	vals := this.values
	n := len(vals)
	return n
}

/*
 * Retrieve all elements from the circular buffer.
 */
func (this *bufferStruct) Retrieve(buf []float64) error {
	values := this.values
	n := len(values)
	m := len(buf)

	/*
	 * Ensure the target buffer is of equal size.
	 */
	if n != m {
		return fmt.Errorf("%s", "Target buffer must be of the same size as source buffer.")
	} else {
		this.mutex.RLock()
		ptr := this.pointer
		tailSize := n - ptr
		copy(buf[0:tailSize], values[ptr:n])
		copy(buf[tailSize:n], values[0:ptr])
		this.mutex.RUnlock()
		return nil
	}

}

/*
 * Creates a circular buffer of a certain size.
 */
func CreateBuffer(size int) Buffer {
	values := make([]float64, size)
	m := sync.RWMutex{}

	/*
	 * Create circular buffer.
	 */
	buf := bufferStruct{
		mutex:   m,
		values:  values,
		pointer: 0,
	}

	return &buf
}
