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

// NewRandomizer return a new Randomizer
func NewRandomizer() Randomizer {
	return &randomize{random: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// randomize is the default Randomizer
type randomize struct {
	sync.Mutex
	random *rand.Rand
}

// Intn implement Randomizer interface
func (r *randomize) Intn(n int) int {
	r.Lock()
	defer r.Unlock()
	return r.random.Intn(n)
}

// Shuffle implement Randomizer interface
func (r *randomize) Shuffle(n int, swap func(i, j int)) {
	r.Lock()
	defer r.Unlock()
	r.random.Shuffle(n, swap)
}
