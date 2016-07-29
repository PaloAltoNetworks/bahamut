package bahamut

type registration struct {
	topic string
	ch    chan *Publication
}

// localPubSub implements a PubSubServer using local channels
type localPubSub struct {
	subscribers  map[string][]chan *Publication
	register     chan *registration
	unregister   chan *registration
	publications chan *Publication
	stop         chan bool
}

// newlocalPubSub returns a new localPubSub.
func newlocalPubSub(services []string) *localPubSub {

	return &localPubSub{
		subscribers:  map[string][]chan *Publication{},
		register:     make(chan *registration, 2),
		unregister:   make(chan *registration, 2),
		stop:         make(chan bool, 2),
		publications: make(chan *Publication, 1024),
	}
}

func (p *localPubSub) registerSubscriberChannel(c chan *Publication, topic string) {

	p.register <- &registration{ch: c, topic: topic}
}

func (p *localPubSub) unregisterSubscriberChannel(c chan *Publication, topic string) {

	p.unregister <- &registration{ch: c, topic: topic}
}

func (p *localPubSub) listen() {

	for {
		select {
		case reg := <-p.register:
			if _, ok := p.subscribers[reg.topic]; !ok {
				p.subscribers[reg.topic] = []chan *Publication{}
			}

			p.subscribers[reg.topic] = append(p.subscribers[reg.topic], reg.ch)

		case reg := <-p.unregister:
			for i, sub := range p.subscribers[reg.topic] {
				if sub == reg.ch {
					p.subscribers[reg.topic] = append(p.subscribers[reg.topic][:i], p.subscribers[reg.topic][i+1:]...)
					close(sub)
					break
				}
			}

		case publication := <-p.publications:
			for _, sub := range p.subscribers[publication.Topic] {
				go func(c chan *Publication) { c <- publication }(sub)
			}

		case <-p.stop:
			p.subscribers = map[string][]chan *Publication{}
			return
		}
	}
}

// Publish publishes a publication.
func (p *localPubSub) Publish(publication *Publication) error {

	p.publications <- publication

	return nil
}

// Subscribe will subscribe the given channel to the given topic
func (p *localPubSub) Subscribe(c chan *Publication, topic string) func() {

	unsubscribe := make(chan bool)

	p.registerSubscriberChannel(c, topic)

	go func() {
		for {
			select {
			case <-unsubscribe:
				p.unregisterSubscriberChannel(c, topic)
				return
			}
		}
	}()

	return func() { unsubscribe <- true }
}

// Connect connects the PubSubServer to the remote service.
func (p *localPubSub) Connect() Waiter {

	abort := make(chan bool, 2)
	connected := make(chan bool, 2)

	go func() {
		go p.listen()
		connected <- true
	}()

	return connectionWaiter{
		ok:    connected,
		abort: abort,
	}
}

// Disconnect disconnects the PubSubServer from the remote service..
func (p *localPubSub) Disconnect() {

	p.stop <- true
}
