package bahamut

import "sync"

type registration struct {
	topic string
	ch    chan *Publication
}

// localPubSub implements a PubSubClient using local channels
type localPubSub struct {
	subscribers  map[string][]chan *Publication
	register     chan *registration
	unregister   chan *registration
	publications chan *Publication
	stop         chan struct{}

	lock *sync.Mutex
}

// newlocalPubSub returns a new localPubSub.
func newlocalPubSub(services []string) *localPubSub {

	return &localPubSub{
		subscribers:  map[string][]chan *Publication{},
		register:     make(chan *registration),
		unregister:   make(chan *registration),
		stop:         make(chan struct{}),
		publications: make(chan *Publication, 1024),
		lock:         &sync.Mutex{},
	}
}

// Publish publishes a publication.
func (p *localPubSub) Publish(publication *Publication) error {

	p.publications <- publication

	return nil
}

// Subscribe will subscribe the given channel to the given topic
func (p *localPubSub) Subscribe(c chan *Publication, errors chan error, topic string, args ...interface{}) func() {

	unsubscribe := make(chan struct{})

	p.registerSubscriberChannel(c, topic)

	go func() {
		<-unsubscribe
		p.unregisterSubscriberChannel(c, topic)
	}()

	return func() { close(unsubscribe) }
}

// Connect connects the PubSubClient to the remote service.
func (p *localPubSub) Connect() Waiter {

	abort := make(chan struct{})
	connected := make(chan bool)

	go func() {
		go p.listen()
		connected <- true
	}()

	return connectionWaiter{
		ok:    connected,
		abort: abort,
	}
}

// Disconnect disconnects the PubSubClient from the remote service..
func (p *localPubSub) Disconnect() error {

	close(p.stop)

	return nil
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
			p.lock.Lock()
			if _, ok := p.subscribers[reg.topic]; !ok {
				p.subscribers[reg.topic] = []chan *Publication{}
			}

			p.subscribers[reg.topic] = append(p.subscribers[reg.topic], reg.ch)
			p.lock.Unlock()

		case reg := <-p.unregister:
			p.lock.Lock()
			for i, sub := range p.subscribers[reg.topic] {
				if sub == reg.ch {
					p.subscribers[reg.topic] = append(p.subscribers[reg.topic][:i], p.subscribers[reg.topic][i+1:]...)
					close(sub)
					break
				}
			}
			p.lock.Unlock()

		case publication := <-p.publications:

			p.lock.Lock()
			for _, sub := range p.subscribers[publication.Topic] {
				go func(s chan *Publication, p *Publication) { s <- p.Duplicate() }(sub, publication)
			}
			p.lock.Unlock()

		case <-p.stop:
			p.lock.Lock()
			p.subscribers = map[string][]chan *Publication{}
			p.lock.Unlock()
			return
		}
	}
}
