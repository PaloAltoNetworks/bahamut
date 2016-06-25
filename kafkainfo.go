// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// KafkaInfo represents Redis connection information
type KafkaInfo struct {
	Addresses []string
	Topic     string
}

// NewKafkaInfo returns a new RedisInfo
func NewKafkaInfo(addresses []string, topic string) *KafkaInfo {

	return &KafkaInfo{
		Addresses: addresses,
		Topic:     topic,
	}
}

func (k *KafkaInfo) makeProducer() sarama.SyncProducer {

	producer, err := sarama.NewSyncProducer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info": k,
		}).Error("unable to create kafka producer")

		return nil
	}

	return producer
}

func (k *KafkaInfo) makeConsumer() sarama.Consumer {

	consumer, err := sarama.NewConsumer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info": k,
		}).Error("unable to create kafka consumer")

		return nil
	}

	return consumer
}

func (k *KafkaInfo) String() string {

	return fmt.Sprintf("<kafkaInfo addresses: %v>", k.Addresses)
}
