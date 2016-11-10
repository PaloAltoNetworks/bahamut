package bahamut

import "sync"

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

	lock *sync.Mutex
}

// newlocalPubSub returns a new localPubSub.
func newlocalPubSub(services []string) *localPubSub {

	return &localPubSub{
		subscribers:  map[string][]chan *Publication{},
		register:     make(chan *registration, 2),
		unregister:   make(chan *registration, 2),
		stop:         make(chan bool, 2),
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
func (p *localPubSub) Subscribe(c chan *Publication, errors chan error, topic string) func() {

	unsubscribe := make(chan bool)

	p.registerSubscriberChannel(c, topic)

	go func() {
		<-unsubscribe
		p.unregisterSubscriberChannel(c, topic)
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
			subs := p.subscribers[reg.topic]
			for i, sub := range subs {
				if sub == reg.ch {
					subs = append(subs[:i], subs[i+1:]...)
					close(sub)
					break
				}
			}
			p.lock.Unlock()

		case publication := <-p.publications:
			p.lock.Lock()
			for _, sub := range p.subscribers[publication.Topic] {
				go func(c chan *Publication) { c <- publication }(sub)
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

func (p *localPubSub) chansForTopic(topic string) []chan *Publication {

	p.lock.Lock()
	defer p.lock.Unlock()
	return p.subscribers[topic]
}
