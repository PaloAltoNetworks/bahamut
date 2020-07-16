package push

type MovingAverage struct {
	samples int
	ring    []float64
	nextIdx int
}

// Average return the average of the samples
// If samples are not compplete it returns 0
func (m *MovingAverage) Average() float64 {

	var sum = float64(0)

	for _, value := range m.ring {
		// we need a complete sample
		if value == 0 {
			return 0
		}
		sum += value
	}

	return sum / float64(len(m.ring))
}

// Add will add a value to the ring.
func (m *MovingAverage) Add(value float64) {

	m.ring[m.nextIdx] = value
	m.nextIdx = (m.nextIdx + 1) % m.samples

}

// NewMovingAverage return a new MovingAverage
func NewMovingAverage(samples int) *MovingAverage {
	return &MovingAverage{
		samples: samples,
		ring:    make([]float64, samples),
		nextIdx: 0,
	}
}
