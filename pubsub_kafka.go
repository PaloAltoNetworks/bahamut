package bahamut

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// kafkaPubSub implements a PubSubServer using Kafka
type kafkaPubSub struct {
	services      []string
	producer      sarama.SyncProducer
	retryInterval time.Duration
}

// newKafkaPubSub Initializes the publishing.
func newKafkaPubSub(services []string) *kafkaPubSub {

	return &kafkaPubSub{
		services:      services,
		retryInterval: 5 * time.Second,
	}
}

// Publish publishes a publication.
func (p *kafkaPubSub) Publish(publication *Publication) error {

	if p.producer == nil {
		return fmt.Errorf("Not connected to kafka. Messages dropped.")
	}

	saramaMsg := &sarama.ProducerMessage{
		Topic: publication.Topic,
		Value: sarama.ByteEncoder(publication.data),
	}

	_, _, err := p.producer.SendMessage(saramaMsg)
	return err
}

// Subscribe will subscribe the given channel to the given topic
func (p *kafkaPubSub) Subscribe(pubs chan *Publication, errs chan error, topic string) func() {
	return p.SubscribeWithOptions(pubs, errs, topic, 0, sarama.OffsetNewest)
}

// Subscribe will subscribe the given channel to the given topic, partition and offset
func (p *kafkaPubSub) SubscribeWithOptions(c chan *Publication, errs chan error, topic string, partition int32, offset int64) func() {
	unsubscribe := make(chan bool)

	go func() {

		defer func() {
			close(c)
			close(errs)
		}()

		var consumer sarama.Consumer
		var partitionConsumer sarama.PartitionConsumer

		for consumer == nil || partitionConsumer == nil {

			var err1, err2 error
			consumer, err1 = sarama.NewConsumer(p.services, nil)

			if err1 == nil {
				partitionConsumer, err2 = consumer.ConsumePartition(topic, partition, offset)
			}

			if err1 == nil && err2 == nil {
				defer func() {
					consumer.Close()
					partitionConsumer.Close()
				}()
				break
			}

			log.WithFields(log.Fields{
				"package":        "bahamut",
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
			case data, ok := <-partitionConsumer.Messages():
				if !ok {
					errs <- fmt.Errorf("kafka partition consumer channel returned empty data")
					continue
				}
				publication := NewPublication(topic)
				publication.data = data.Value
				c <- publication
			case err := <-partitionConsumer.Errors():
				errs <- err
				return
			case <-unsubscribe:
				return
			}
		}
	}()

	return func() { unsubscribe <- true }
}

// Connect connects the PubSubServer to the remote service.
func (p *kafkaPubSub) Connect() Waiter {

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
func (p *kafkaPubSub) Disconnect() {

	if p.producer != nil {
		if err := p.producer.Close(); err != nil {
			log.WithFields(log.Fields{
				"package": "bahamut",
				"error":   err,
			}).Error("Unable to close to kafka producer.")
		}

		p.producer = nil
	}
}
