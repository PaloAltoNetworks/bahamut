// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// PushServerConfig represents Redis connection information
type PushServerConfig struct {
	Addresses       []string
	DefaultTopic    string
	Authorizer      Authorizer
	Authenticator   Authenticator
	sessionsHandler PushSessionsHandler
}

// MakePushServerConfig returns a new RedisInfo
func MakePushServerConfig(addresses []string, defaultTopic string, sessionsHandler PushSessionsHandler) PushServerConfig {

	if len(addresses) > 0 && defaultTopic == "" {
		panic("you must pass a default topic if you provide kafka addresses")
	}

	if len(addresses) == 0 && defaultTopic != "" {
		panic("you must pass at least one kafka address if you provide default topic")
	}

	return PushServerConfig{
		Addresses:       addresses,
		DefaultTopic:    defaultTopic,
		sessionsHandler: sessionsHandler,
	}
}

func (k PushServerConfig) makeProducer() sarama.SyncProducer {

	producer, err := sarama.NewSyncProducer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info":  k,
			"error": err,
		}).Error("unable to create kafka producer")

		return nil
	}

	return producer
}

func (k PushServerConfig) makeConsumer() sarama.Consumer {

	consumer, err := sarama.NewConsumer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info":  k,
			"error": err,
		}).Error("unable to create kafka consumer")

		return nil
	}

	return consumer
}

// HasKafka returns true is the PushServerConfig has Kafka server information
func (k PushServerConfig) HasKafka() bool {

	return len(k.Addresses) > 0
}

func (k PushServerConfig) String() string {

	return fmt.Sprintf("<PushServerConfig Addresses: %v DefaultTopic: %s>", k.Addresses, k.DefaultTopic)
}
