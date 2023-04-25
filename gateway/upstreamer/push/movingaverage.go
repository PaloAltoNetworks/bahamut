package push

import "fmt"

// MovingAverage represent a moving average
// give a sample size.
type movingAverage struct {
	ring             []float64
	sampleSize       int
	nextIdx          int
	samplingComplete bool
}

// newMovingAverage return a new movingAverage
func newMovingAverage(sampleSize int) movingAverage {

	if sampleSize <= 0 {
		panic("sampleSize must be greather than 0.")
	}

	return movingAverage{
		sampleSize: sampleSize,
		ring:       make([]float64, sampleSize),
	}
}

// average return the average of the sampleSize
// If sampleSize are not compplete it returns 0
func (m movingAverage) average() (float64, error) {

	sum := .0

	if !m.samplingComplete {
		return sum, fmt.Errorf("cannot compute average without a full sampling")
	}

	for _, value := range m.ring {
		sum += value
	}

	return sum / float64(m.sampleSize), nil
}

// append will insert a new value to the ring and return a copy
// of itself
func (m movingAverage) append(value float64) movingAverage {

	nm := newMovingAverage(m.sampleSize)
	nm.samplingComplete = m.samplingComplete
	for i := range m.ring {
		nm.ring[i] = m.ring[i]
	}

	nm.ring[m.nextIdx] = value
	nm.nextIdx = (m.nextIdx + 1) % nm.sampleSize
	if nm.nextIdx == 0 {
		nm.samplingComplete = true
	}

	return nm
}
