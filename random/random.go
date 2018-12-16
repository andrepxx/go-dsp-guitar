package random

/*
 * Interface type for a pseudo random number generator.
 */
type PseudoRandomNumberGenerator interface {
	NextFloat() float64
}

/*
 * Data structure representing a linear congruency generator.
 */
type linearCongruencyGenerator struct {
	a uint64
	b uint64
	n uint64
	x uint64
}

/*
 * Samples a new random number in the interval [0, 1] from a uniform distribution.
 */
func (this *linearCongruencyGenerator) NextFloat() float64 {
	a := this.a
	b := this.b
	n := this.n
	x := this.x
	x = ((a * x) + b) % n
	this.x = x
	xFloat := float64(x)
	xMax := n - 1
	xMaxFloat := float64(xMax)
	result := xFloat / xMaxFloat
	return result
}

/*
 * Creates a new pseudo random number generator.
 */
func CreatePRNG(seed uint64) PseudoRandomNumberGenerator {
	n := uint64((1 << 31) - 1)
	x := ((64979 * seed) + 83) % n

	/*
	 * Initialize a new LCG.
	 */
	generator := linearCongruencyGenerator{
		a: 16807,
		b: 0,
		n: n,
		x: x,
	}

	return &generator
}
