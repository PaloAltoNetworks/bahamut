package push

import (
	"math/rand"
	"sync"
	"time"
)

// A Randomizer reprensents an interface to randomize
type Randomizer interface {
	Intn(int) int
	Shuffle(n int, swap func(i, j int))
}

// newRandomizer return a new Randomizer
func newRandomizer() Randomizer {
	return &defaultRandomizer{random: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// defaultRandomizer is the default Randomizer
type defaultRandomizer struct {
	sync.Mutex
	random *rand.Rand
}

// Intn implement Randomizer interface
func (r *defaultRandomizer) Intn(n int) int {
	r.Lock()
	defer r.Unlock()
	return r.random.Intn(n)
}

// Shuffle implement Randomizer interface
func (r *defaultRandomizer) Shuffle(n int, swap func(i, j int)) {
	r.Lock()
	r.random.Shuffle(n, swap)
	r.Unlock()
}
