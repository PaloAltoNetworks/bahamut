package pubsub

import (
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/aporeto-inc/bahamut/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKafka_NewPubSubServer(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newKafkaPubSub([]string{"123:123"})

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.services[0], ShouldEqual, "123:123")
		})
	})
}

func TestKafka_StartStop(t *testing.T) {

	Convey("Given I create a new PubSubServer with a bad kafka address", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
		})
		defer broker.Close()

		ps := newKafkaPubSub([]string{})
		ps.retryInterval = 1 * time.Millisecond

		Convey("When I start the server", func() {

			connected := ps.Connect().Wait(2 * time.Millisecond)

			Convey("Then the producer should be nil", func() {
				So(connected, ShouldBeFalse)
			})
		})
	})

	Convey("Given I create a new PubSubServer with a good kafka address", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
		})
		defer broker.Close()

		ps := newKafkaPubSub([]string{broker.Addr()})

		Convey("When I start the server", func() {

			ps.Connect().Wait(300 * time.Millisecond)
			defer ps.Disconnect()

			Convey("Then the producer should not be nil", func() {
				So(ps.producer, ShouldNotBeNil)
			})

			Convey("When I stop the server", func() {

				ps.Disconnect()

				Convey("Then the producer should be closed", func() {
					So(ps.producer, ShouldBeNil)
				})
			})
		})
	})
}

func TestKafka_Publish(t *testing.T) {

	Convey("Given I try to publish while not connected", t, func() {

		ps := newKafkaPubSub([]string{})

		Convey("When I publish something", func() {

			list := mock.NewList()
			list.Name = "l1"
			list.ID = "xxx"

			publication := NewPublication("topic")
			publication.Encode(list)

			err := ps.Publish(publication)

			Convey("Then error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I start a PubSubServer with a good kafka address", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
			"ProduceRequest": sarama.NewMockProduceResponse(t),
		})
		defer broker.Close()

		ps := newKafkaPubSub([]string{broker.Addr()})
		ps.Connect().Wait(300 * time.Millisecond)

		defer ps.Disconnect()

		Convey("When I publish something", func() {

			list := mock.NewList()
			list.Name = "l1"
			list.ID = "xxx"

			publication := NewPublication("topic")
			publication.Encode(list)

			err := ps.Publish(publication)

			Convey("Then error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given I start a PubSubServer with a good kafka address but I can't produce", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
			"ProduceRequest": sarama.NewMockProduceResponse(t).
				SetError("topic", 0, sarama.ErrBrokerNotAvailable),
		})
		defer broker.Close()

		ps := newKafkaPubSub([]string{broker.Addr()})
		ps.Connect().Wait(300 * time.Millisecond)

		defer ps.Disconnect()

		Convey("When I publish something", func() {

			publication := NewPublication("topic")

			err := ps.Publish(publication)

			Convey("Then the error should not be nil. ", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestKafka_Subscribe(t *testing.T) {

	Convey("Given I try to subscribe but I cannot connect", t, func() {

		ps := newKafkaPubSub([]string{})
		ps.retryInterval = 1 * time.Millisecond

		Convey("When I subscribe to something", func() {

			c := make(chan *Publication)
			u := ps.Subscribe(c, "topic")
			<-time.After(2 * time.Millisecond)

			Convey("Then error it should retry until I unsubscribe", func() {
				u()
			})
		})
	})

	Convey("Given I try to subscribe", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
			"FetchRequest": sarama.NewMockFetchResponse(t, 1).
				SetMessage("topic", 0, 0, sarama.StringEncoder("hello")),
		})
		defer broker.Close()

		ps := newKafkaPubSub([]string{broker.Addr()})

		Convey("When I subscribe", func() {

			c := make(chan *Publication)
			u := ps.Subscribe(c, "topic")

			Convey("Then unsubscribe channel should not be nil", func() {
				So(u, ShouldNotBeNil)
			})

			Convey("When I read from the subscription channel", func() {
				p := <-c

				Convey("Then my channel should receive a publication", func() {
					So(string(p.Data()), ShouldEqual, "hello")
				})

				Convey("When I use the unsubscribe channel", func() {
					u()
					_, ok := <-c

					Convey("Then my channel should be closed", func() {
						So(ok, ShouldBeFalse)
					})
				})
			})
		})
	})
}
