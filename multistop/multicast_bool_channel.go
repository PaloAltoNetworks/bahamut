package multistop

// A MultiStop is a struct that allows
// to send a boolean that will be sent to all registered
// channel
type MultiStop struct {
	channels []chan bool
}

// NewMultiStop returns a new MultiCastBooleanChannel.
func NewMultiStop() *MultiStop {

	return &MultiStop{
		channels: []chan bool{},
	}
}

// Register registers the given channel to the multicast.
func (p *MultiStop) Register(ch chan bool) {

	p.channels = append(p.channels, ch)
}

// Unregister unregisters the given channel from the multicast.
func (p *MultiStop) Unregister(ch chan bool) {

	var i int
	var c chan bool

	for i, c = range p.channels {
		if c == ch {
			p.channels = append(p.channels[:i], p.channels[i+1:]...)
			break
		}
	}
}

// Send sends the given boolean to all registered channels
func (p *MultiStop) Send(b bool) {

	for _, c := range p.channels {
		c <- b
	}
}
