package pubsub

// A Server is a structure that provides publish subscribe mechanism.
type Server interface {
	Publish(publications ...*Publication) error
	Subscribe(c chan *Publication, topic string) func()
	Start()
	Stop()
}

// NewServer Initializes the PubSubServer.
func NewServer(services []string) Server {

	return newKafkaPubSubServer(services)
}
