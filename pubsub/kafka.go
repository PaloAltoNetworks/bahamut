package pubsub

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/bahamut/multicaststop"
)

// kafkaPubSubServer implements a PubSubServer using Kafka
type kafkaPubSubServer struct {
	services      []string
	producer      sarama.SyncProducer
	publications  chan *Publication
	retryInterval time.Duration
	multicast     *multicaststop.MultiCastBooleanChannel
}

// newPubSubServer Initializes the publishing.
func newKafkaPubSubServer(services []string) *kafkaPubSubServer {

	return &kafkaPubSubServer{
		services:      services,
		publications:  make(chan *Publication, 1024),
		multicast:     multicaststop.NewMultiCastBooleanChannel(),
		retryInterval: 5 * time.Second,
	}
}

// listen listens from the channel, and publishes messages to kafka
func (p *kafkaPubSubServer) listen() {

	stopCh := make(chan bool)
	p.multicast.Register(stopCh)

	defer p.multicast.Unregister(stopCh)

	for {
		select {
		case publication := <-p.publications:

			saramaMsg := &sarama.ProducerMessage{
				Topic: publication.Topic,
				Value: sarama.ByteEncoder(publication.data),
			}

			if _, _, err := p.producer.SendMessage(saramaMsg); err != nil {
				log.WithFields(log.Fields{
					"publication": saramaMsg,
					"materia":     "bahamut",
				}).Warn("Unable to publish message to Kafka. Message dropped.")
			}

		case <-stopCh:
			return
		}
	}
}

// Publish sends multiple messages. Creates the message and puts it
// in the queue, but doesn't wait for this to be transmitted
func (p *kafkaPubSubServer) Publish(publications ...*Publication) error {

	if p.producer == nil {
		return fmt.Errorf("Not connected to kafka. Messages dropped.")
	}

	for _, publication := range publications {
		select {
		case p.publications <- publication:
		default:
			return fmt.Errorf("Queue is full. Messages dropped.")
		}
	}

	return nil
}

// Subscribe will subscribe the given channel to the given topic
func (p *kafkaPubSubServer) Subscribe(c chan *Publication, topic string) func() {

	unsubscribe := make(chan bool)

	go func() {

		defer func() {
			close(c)
		}()

		var consumer sarama.Consumer
		var partition sarama.PartitionConsumer

		for consumer == nil || partition == nil {

			var err1, err2 error
			consumer, err1 = sarama.NewConsumer(p.services, nil)

			if err1 == nil {
				partition, err2 = consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
			}

			if err1 == nil && err2 == nil {
				break
			}

			log.WithFields(log.Fields{
				"materia":        "bahamut",
				"topic":          topic,
				"consumerError":  err1,
				"partitionError": err2,
			}).Warn("Unable to create paritition consumer. Retrying in 5 seconds...")

			select {
			case <-time.After(p.retryInterval):
			case <-unsubscribe:
				return
			}
		}

		for {
			select {
			case data := <-partition.Messages():
				publication := NewPublication(topic)
				publication.data = data.Value
				c <- publication
			case <-unsubscribe:
				return
			}
		}
	}()

	return func() { unsubscribe <- true }
}

// Start starts the publisher
func (p *kafkaPubSubServer) Start() {

	stopCh := make(chan bool)
	p.multicast.Register(stopCh)

	defer func() {
		if p.producer != nil {
			p.producer.Close()
			p.producer = nil
		}
	}()

	for p.producer == nil {

		var err error
		p.producer, err = sarama.NewSyncProducer(p.services, nil)

		if err == nil {
			break
		}

		log.WithFields(log.Fields{
			"services": p.services,
		}).Warn("Unable to create to kafka producer retrying in 5 seconds.")

		select {
		case <-time.After(p.retryInterval):
			continue
		case <-stopCh:
			p.multicast.Unregister(stopCh)
			return
		}
	}

	p.multicast.Unregister(stopCh)
	p.listen()
}

// Stop stops the publishing.
func (p *kafkaPubSubServer) Stop() {

	p.multicast.Send(true)
}
