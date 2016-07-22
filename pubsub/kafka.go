package pubsub

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// kafkaPubSubServer implements a PubSubServer using Kafka
type kafkaPubSubServer struct {
	services      []string
	producer      sarama.SyncProducer
	retryInterval time.Duration
}

// newPubSubServer Initializes the publishing.
func newKafkaPubSubServer(services []string) *kafkaPubSubServer {

	return &kafkaPubSubServer{
		services:      services,
		retryInterval: 5 * time.Second,
	}
}

// Publish publishes a publication.
func (p *kafkaPubSubServer) Publish(publication *Publication) error {

	if p.producer == nil {
		return fmt.Errorf("Not connected to kafka. Messages dropped.")
	}

	saramaMsg := &sarama.ProducerMessage{
		Topic: publication.Topic,
		Value: sarama.ByteEncoder(publication.data),
	}

	if _, _, err := p.producer.SendMessage(saramaMsg); err != nil {
		return err
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
				"retryIn":        p.retryInterval,
			}).Warn("Unable to create partition consumer. Retrying...")

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

// Connect connects the PubSubServer to the remote service.
func (p *kafkaPubSubServer) Connect() Waiter {

	abort := make(chan bool, 2)
	connected := make(chan bool, 2)

	go func() {
		for p.producer == nil {

			var err error
			p.producer, err = sarama.NewSyncProducer(p.services, nil)

			if err == nil {
				break
			}

			log.WithFields(log.Fields{
				"services": p.services,
				"package":  "bahamut",
				"retryIn":  p.retryInterval,
			}).Warn("Unable to create to kafka producer retrying in 5 seconds.")

			select {
			case <-time.After(p.retryInterval):
			case <-abort:
				connected <- false
				return
			}
		}
		connected <- true
	}()

	return connectionWaiter{
		ok:    connected,
		abort: abort,
	}
}

// Disconnect disconnects the PubSubServer from the remote service..
func (p *kafkaPubSubServer) Disconnect() {

	if p.producer != nil {
		p.producer.Close()
		p.producer = nil
	}
}
