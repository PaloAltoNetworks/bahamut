package push

import (
	"fmt"
	"sync"
	"time"
)

// responseTimes holds the endpoints response times
type responseTimes struct {
	responseTimeSamples int
	data                sync.Map
}

// newResponseTimes return a new responseTimes
func newResponseTimes(samples int) *responseTimes {
	return &responseTimes{responseTimeSamples: samples}
}

// StoreResponseTime implement the ResponseTimeCollector interface to add new
// samples into the responseTimeSamples sync map
func (r *responseTimes) StoreResponseTime(address string, responseTime time.Duration) {

	if values, ok := r.data.Load(address); ok {
		values.(*movingAverage).insertValue(float64(responseTime.Microseconds()))
	} else {
		r.data.Store(address, newMovingAverage(r.responseTimeSamples))
		r.StoreResponseTime(address, responseTime)
	}
}

// deleteResponseTimes delete the reponseTimes for a give address
func (r *responseTimes) deleteResponseTimes(address string) {
	r.data.Delete(address)
}

// getResponseTime return the average responsetime for a give address
func (r *responseTimes) getResponseTime(address string) (float64, error) {

	if ma, ok := r.data.Load(address); ok {
		v, err := ma.(*movingAverage).average()
		if err != nil {
			return 0, err
		}
		return v, nil
	}

	return 0, fmt.Errorf("Response time is not tracked for %v", address)
}

// MovingAverage represent a moving average
// give an number of samples.
type movingAverage struct {
	samples          int
	ring             []float64
	nextIdx          int
	samplingComplete bool
}

// newMovingAverage return a new movingAverage
func newMovingAverage(samples int) *movingAverage {
	return &movingAverage{
		samples: samples,
		ring:    make([]float64, samples),
		nextIdx: 0,
	}
}

// average return the average of the samples
// If samples are not compplete it returns 0
func (m *movingAverage) average() (float64, error) {

	var sum = .0

	if !m.samplingComplete {
		return sum, fmt.Errorf("Cannot compute average without a full sampling")
	}

	for _, value := range m.ring {
		sum += value
	}

	return sum / float64(len(m.ring)), nil
}

// insertValue will insert a new value to the ring.
func (m *movingAverage) insertValue(value float64) {

	m.ring[m.nextIdx] = value
	m.nextIdx = (m.nextIdx + 1) % m.samples
	if m.nextIdx == 0 {
		m.samplingComplete = true
	}
}
