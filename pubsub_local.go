// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"context"
	"sync"
)

type registration struct {
	ch    chan *Publication
	topic string
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

// NewLocalPubSubClient returns a PubSubClient backed by local channels.
func NewLocalPubSubClient() PubSubClient {

	return newlocalPubSub()
}

// newlocalPubSub returns a new localPubSub.
func newlocalPubSub() *localPubSub {

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
func (p *localPubSub) Publish(publication *Publication, opts ...PubSubOptPublish) error {

	p.publications <- publication

	return nil
}

// Subscribe will subscribe the given channel to the given topic
func (p *localPubSub) Subscribe(c chan *Publication, errors chan error, topic string, opts ...PubSubOptSubscribe) func() {

	unsubscribe := make(chan struct{})

	p.registerSubscriberChannel(c, topic)

	go func() {
		<-unsubscribe
		p.unregisterSubscriberChannel(c, topic)
	}()

	return func() { close(unsubscribe) }
}

// Connect connects the PubSubClient to the remote service.
func (p *localPubSub) Connect(ctx context.Context) error {

	go p.listen()

	return nil
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
			var wg sync.WaitGroup
			for _, sub := range p.subscribers[publication.Topic] {
				wg.Add(1)
				go func(s chan *Publication, p *Publication) {
					defer wg.Done()
					s <- p.Duplicate()
				}(sub, publication)
			}
			wg.Wait()
			p.lock.Unlock()

		case <-p.stop:
			p.lock.Lock()
			p.subscribers = map[string][]chan *Publication{}
			p.lock.Unlock()
			return
		}
	}
}
