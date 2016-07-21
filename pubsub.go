package bahamut

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// Publication is a structure that can be published to a PublishServer.
type Publication struct {
	data  []byte
	Topic string
}

// NewPublication returns a new Publication.
func NewPublication(topic string) *Publication {

	return &Publication{
		Topic: topic,
	}
}

// Encode the given object into the publication.
func (p *Publication) Encode(o interface{}) error {

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(o); err != nil {
		return err
	}

	p.data = buffer.Bytes()

	return nil
}

// Data returns the raw data contained in the publication.
func (p *Publication) Data() []byte {

	return p.data
}

// Decode decodes the data into the given dest.
func (p *Publication) Decode(dest interface{}) error {

	if err := json.NewDecoder(bytes.NewReader(p.data)).Decode(&dest); err != nil {
		return err
	}

	return nil
}

// publishServer holds configuration
type publishServer struct {
	services     []string
	producer     sarama.SyncProducer
	consumer     sarama.Consumer
	publications chan *Publication
	close        chan bool
}

// newPublishServer Initializes the publishing.
func newPublishServer(services []string) *publishServer {

	return &publishServer{
		services:     services,
		publications: make(chan *Publication, 1024),
		close:        make(chan bool),
	}
}

// listen listens from the channel, and publishes messages to kafka
func (p *publishServer) listen() {

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

		case <-p.close:
			p.producer.Close()
			break
		}
	}
}

// Publish sends multiple messages. Creates the message and puts it
// in the queue, but doesn't wait for this to be transmitted
func (p *publishServer) Publish(publications ...*Publication) {

	if p.producer == nil {
		log.WithFields(log.Fields{
			"materia": "bahamut",
		}).Warn("Not connected to kafka. Messages dropped.")

		return
	}

	for _, publication := range publications {
		select {
		case p.publications <- publication:
		default:
			log.WithFields(log.Fields{
				"materia": "bahamut",
			}).Error("Queue is full. Messages dropped.")
		}
	}
}

// Subscribe will subscribe the given channel to the given topic
func (p *publishServer) Subscribe(c chan *Publication, t string) chan bool {

	unsubscribe := make(chan bool)

	go func() {
		var partition sarama.PartitionConsumer
		var err error

		for partition == nil {

			if partition, err = p.consumer.ConsumePartition(t, 0, sarama.OffsetNewest); err != nil {
				break
			}

			log.WithFields(log.Fields{
				"materia": "bahamut",
			}).Warn("Unable to create consumer. Retrying in 5 seconds...")

			select {
			case <-time.After(5 * time.Second):
				continue
			case <-p.close:
				return
			}
		}

		for {
			select {
			case data := <-partition.Messages():
				publication := NewPublication(t)
				publication.data = data.Value
				c <- publication
			case <-p.close:
				return
			case <-unsubscribe:
				return
			}
		}
	}()

	return unsubscribe
}

// Start starts the publisher
func (p *publishServer) start() {

	for p.producer == nil {

		var perr, cerr error
		p.producer, perr = sarama.NewSyncProducer(p.services, nil)
		p.consumer, cerr = sarama.NewConsumer(p.services, nil)

		if perr == nil && cerr == nil {
			break
		}

		log.WithFields(log.Fields{
			"services": p.services,
		}).Warn("Unable to create to kafka producer retrying in 5 seconds.")

		<-time.After(5 * time.Second)
	}

	p.listen()
}

// stop stops the publishing.
func (p *publishServer) stop() {

	p.close <- true
}
