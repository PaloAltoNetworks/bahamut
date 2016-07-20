// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	"github.com/Shopify/sarama"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKakfaInfo_MakePushServerConfig(t *testing.T) {

	Convey("Given I create have a new config", t, func() {

		config := MakePushServerConfig([]string{":1234"}, "topic", nil)

		Convey("Then the kafka info should have the address set", func() {
			So(len(config.kafkaAddresses), ShouldEqual, 1)
			So(config.kafkaAddresses[0], ShouldEqual, ":1234")
		})

		Convey("Then the kafka info should have the default topic set", func() {
			So(config.defaultTopic, ShouldEqual, "topic")
		})

		Convey("Then enabled flag should be set", func() {
			So(config.enabled, ShouldBeTrue)
		})
	})

	Convey("Given I create have a new config with an empty address array", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { MakePushServerConfig([]string{}, "topic", nil) }, ShouldPanic)
		})
	})

	Convey("Given I create have a new config with an empty topic", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { MakePushServerConfig([]string{":1234"}, "", nil) }, ShouldPanic)
		})
	})
}

func TestKakfaInfo_String(t *testing.T) {

	Convey("Given I create have a new config with kafka info", t, func() {

		config := MakePushServerConfig([]string{"127.0.0.1:1234", "127.0.0.1:1235"}, "topic", nil)

		Convey("Then the string representation should be correct", func() {
			So(config.String(), ShouldEqual, "<PushServerConfig Addresses: [127.0.0.1:1234 127.0.0.1:1235] DefaultTopic: topic>")
		})
	})
}

func TestKakfaInfo_HasKafka(t *testing.T) {

	Convey("Given I create have a new config with kafka info", t, func() {

		config := MakePushServerConfig([]string{"127.0.0.1:1234", "127.0.0.1:1235"}, "topic", nil)

		Convey("Then HasKafka should return true", func() {
			So(config.hasKafka(), ShouldBeTrue)
		})
	})

	Convey("Given I create have a new config without kafka info", t, func() {

		config := MakePushServerConfig([]string{}, "", nil)

		Convey("Then HasKafka should return false", func() {
			So(config.hasKafka(), ShouldBeFalse)
		})
	})
}

func TestKakfaInfo_makeProducer(t *testing.T) {

	Convey("Given I create have a new config with a kafka server listen", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		metadataResponse := new(sarama.MetadataResponse)
		metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
		metadataResponse.AddTopicPartition("topic", 0, broker.BrokerID(), nil, nil, sarama.ErrNoError)
		broker.Returns(metadataResponse)
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)

		Convey("When I make a producer", func() {

			p, err := config.makeProducer()

			Convey("Then the errr should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the producer should be correctly set", func() {
				So(p, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create have a new config with no kafka server listen", t, func() {

		config := MakePushServerConfig([]string{":1234"}, "topic", nil)

		Convey("When I make a producer", func() {

			p, err := config.makeProducer()

			Convey("Then the error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the producer should be nil", func() {
				So(p, ShouldBeNil)
			})
		})
	})
}

func TestKakfaInfo_makeConsumer(t *testing.T) {

	Convey("Given I create have a new config with a kafka server listen", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		metadataResponse := new(sarama.MetadataResponse)
		metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
		metadataResponse.AddTopicPartition("topic", 0, broker.BrokerID(), nil, nil, sarama.ErrNoError)
		broker.Returns(metadataResponse)
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)

		Convey("When I make a consumer", func() {

			p, err := config.makeConsumer()

			Convey("Then the err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the consumer should be correctly set", func() {
				So(p, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create have a new config with no kafka server listen", t, func() {

		config := MakePushServerConfig([]string{":1234"}, "topic", nil)

		Convey("When I make a producer", func() {

			p, err := config.makeConsumer()

			Convey("Then the err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the consumer should be nil", func() {
				So(p, ShouldBeNil)
			})
		})
	})
}
