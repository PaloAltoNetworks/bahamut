// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"

	"github.com/Shopify/sarama"
)

// A PushServerConfig contains the configuration for the Bahamut Push Server.
type PushServerConfig struct {
	kafkaAddresses  []string
	defaultTopic    string
	sessionsHandler PushSessionsHandler
	enabled         bool
}

// MakePushServerConfig returns a new PushServerConfig.
//
// addresses represents the list of optional Kafka server endpoint to use in order to
// enable the distributed push mechanism. If it empty, Bahamut will only push to local session.
//
// defaultTopic defines the default kafka topic to use. This is used only if addresses are set.
//
// sessionsHandler is an optional object that implements the PushSessionsHandler.
func MakePushServerConfig(addresses []string, defaultTopic string, sessionsHandler PushSessionsHandler) PushServerConfig {

	if len(addresses) > 0 && defaultTopic == "" {
		panic("you must pass a default topic if you provide kafka addresses")
	}

	if len(addresses) == 0 && defaultTopic != "" {
		panic("you must pass at least one kafka address if you provide default topic")
	}

	return PushServerConfig{
		kafkaAddresses:  addresses,
		defaultTopic:    defaultTopic,
		sessionsHandler: sessionsHandler,
		enabled:         true,
	}
}

func (k PushServerConfig) makeProducer() (sarama.SyncProducer, error) {

	producer, err := sarama.NewSyncProducer(k.kafkaAddresses, nil)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func (k PushServerConfig) makeConsumer() (sarama.Consumer, error) {

	consumer, err := sarama.NewConsumer(k.kafkaAddresses, nil)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (k PushServerConfig) hasKafka() bool {

	return len(k.kafkaAddresses) > 0
}

func (k PushServerConfig) String() string {

	return fmt.Sprintf("<PushServerConfig Addresses: %v DefaultTopic: %s>", k.kafkaAddresses, k.defaultTopic)
}
