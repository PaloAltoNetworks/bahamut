package bahamut

import (
	"bytes"
	"encoding/json"
)

// Publication is a structure that can be published to a PublishServer.
type Publication struct {
	data      []byte
	Topic     string
	Partition int32
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

	return json.NewDecoder(bytes.NewReader(p.data)).Decode(&dest)
}
