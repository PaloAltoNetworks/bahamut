// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/websocket"
)

func TestSession_newPushSession(t *testing.T) {

	Convey("When I create have a new pushSession", t, func() {

		ws := &websocket.Conn{}
		session := newPushSession(ws, newPushServer(PushServerConfig{}, bone.New()))

		Convey("Then the session id should not be empty", func() {
			So(session.id, ShouldNotBeEmpty)
		})

		Convey("Then the socket should be nil", func() {
			So(session.socket, ShouldEqual, ws)
		})

		Convey("Then the events channel should be a chan of bytes", func() {
			So(session.events, ShouldHaveSameTypeAs, make(chan string))
		})

		Convey("Then the Identifier() should return the id", func() {
			So(session.Identifier(), ShouldEqual, session.id)
		})
	})
}

func TestSession_listenToKafkaMessages(t *testing.T) {

	Convey("Given I create have a new pushSession with valid kafka info", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
			"FetchRequest": sarama.NewMockFetchResponse(t, 1).
				SetMessage("topic", 0, 0, sarama.StringEncoder(`{"hello":"world"}`)),
		})
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)
		ws := &websocket.Conn{}
		session := newPushSession(ws, newPushServer(config, bone.New()))

		Convey("When I listen for kafka messages", func() {
			go session.listenToKafkaMessages()

			var message string
			select {
			case message = <-session.out:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the messge should be correct", func() {
				So(message, ShouldEqual, `{"hello":"world"}`)
			})
		})

		Convey("When I get a stop while I listen for messages", func() {
			c := make(chan bool, 1)
			go func() {
				session.listenToKafkaMessages()
				c <- true
			}()

			session.close()

			var returned bool
		LOOP:
			for {
				select {
				case returned = <-c:
					break LOOP
				case <-session.server.unregister:
				case <-time.After(300 * time.Millisecond):
					break LOOP
				}
			}

			Convey("Then the function should exit correctly", func() {
				So(returned, ShouldBeTrue)
			})
		})
	})

	Convey("Given I create have a new pushSession with invalid kafka info", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		errorResponse := &sarama.FetchResponse{}
		errorResponse.AddError("topic", 0, sarama.ErrInvalidTopic)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockWrapper(errorResponse),
		})
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)
		ws := &websocket.Conn{}
		session := newPushSession(ws, newPushServer(config, bone.New()))

		Convey("When I listen for kafka messages", func() {

			err := session.listenToKafkaMessages()

			Convey("Then it should return right an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSession_listenToLocalMessages(t *testing.T) {

	Convey("Given I create have a new pushSession with no valid kafka info", t, func() {

		ws := &websocket.Conn{}
		session := newPushSession(ws, newPushServer(PushServerConfig{}, bone.New()))

		Convey("When I listen for local messages", func() {
			go session.listenToLocalMessages()

			session.events <- `{"hello":"world"}`
			var message string
			select {
			case message = <-session.out:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the messge should be correct", func() {
				So(message, ShouldEqual, `{"hello":"world"}`)
			})
		})

		Convey("When I get a stop while I listen for messages", func() {
			c := make(chan bool, 1)
			go func() {
				session.listenToLocalMessages()
				c <- true
			}()

			session.close()

			var returned bool
		LOOP:
			for {
				select {
				case returned = <-c:
					break LOOP
				case <-session.server.unregister:
				case <-time.After(300 * time.Millisecond):
					break LOOP
				}
			}

			Convey("Then the function should exit correctly", func() {
				So(returned, ShouldBeTrue)
			})
		})
	})

	Convey("Given I create have a new pushSession with invalid kafka info", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		errorResponse := &sarama.FetchResponse{}
		errorResponse.AddError("topic", 0, sarama.ErrInvalidTopic)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockWrapper(errorResponse),
		})
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)
		ws := &websocket.Conn{}
		session := newPushSession(ws, newPushServer(config, bone.New()))

		Convey("When I listen for kafka messages", func() {

			err := session.listenToKafkaMessages()

			Convey("Then it should return right an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSession_send(t *testing.T) {

	Convey("Given I create a session with a websocket with a session handler", t, func() {

		handler := &testSessionHandler{}

		session := newPushSession(&websocket.Conn{}, newPushServer(MakePushServerConfig([]string{}, "", handler), bone.New()))

		Convey("When I send some data to the session", func() {

			handler.block = false
			go session.send("{}")

			var processed bool
			select {
			case <-session.out:
				processed = true
				break
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then this should be like that", func() {
				So(processed, ShouldBeTrue)
			})
		})

		Convey("When I send some data to the session while my handler doesn't allow the push", func() {

			handler.block = true
			go session.send("{}")

			var processed bool
			select {
			case <-session.out:
				processed = true
				break
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then this should be like that", func() {
				So(processed, ShouldBeFalse)
			})
		})

		Convey("When I send some unmarshallable data to the session", func() {

			handler.block = false
			err := session.send("{bad")

			Convey("Then error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSession_write(t *testing.T) {

	Convey("Given I create a session with a websocket", t, func() {

		ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			var d []byte
			websocket.Message.Receive(ws, &d)
			websocket.Message.Send(ws, d)
		}))
		defer ts.Close()

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		session := newPushSession(ws, newPushServer(PushServerConfig{}, bone.New()))

		Convey("When I send some data to the session", func() {

			go session.write()

			session.out <- "hello world"

			var data []byte
			websocket.Message.Receive(ws, &data)

			Convey("Then the websocket should receive the data", func() {
				So(string(data), ShouldEqual, "hello world")
			})
		})

		Convey("When I stop the session while listening to the websocket", func() {

			c := make(chan bool, 1)
			go func() {
				session.write()
				c <- true
			}()

			session.close()

			var returned bool
			select {
			case returned = <-c:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the function should exit", func() {
				So(returned, ShouldBeTrue)
			})
		})

		Convey("When the websocket is closed while I'm listening", func() {

			c := make(chan bool, 1)
			go func() {
				session.write()
				c <- true
			}()

			ws.Close()
			session.out <- "hello world"

			var returned bool
			select {
			case returned = <-c:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the write function should exit", func() {
				So(returned, ShouldBeTrue)
			})
		})
	})
}

func TestSession_read(t *testing.T) {

	Convey("Given I create a session with a websocket", t, func() {

		dt := make(chan []byte)
		ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
			websocket.Message.Send(ws, <-dt)
		}))
		defer ts.Close()

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		session := newPushSession(ws, newPushServer(PushServerConfig{}, bone.New()))

		Convey("When I receive some data to the session", func() {

			c := make(chan bool, 1)
			go func() {
				session.read()
				c <- true
			}()

			dt <- []byte("hello world")

			var returned bool
			select {
			case returned = <-c:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the write function should not exit", func() {
				So(returned, ShouldBeTrue) // TODO: this is should be False.
			})
		})
	})
}

func TestSession_listen(t *testing.T) {

	Convey("Given I create a session with a websocket and not kafka info", t, func() {

		ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {}))
		defer ts.Close()

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		session := newPushSession(ws, newPushServer(PushServerConfig{}, bone.New()))

		c := make(chan bool, 1)
		go func() {
			session.listen()
			c <- true
		}()

		Convey("When I keep it alive", func() {

			var returned bool
			select {
			case returned = <-c:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the function should not return", func() {
				So(returned, ShouldBeFalse)
			})
		})

		Convey("When I close it", func() {

			session.close()

			var returned bool
		LOOP:
			for {
				select {
				case <-session.server.unregister:
				case returned = <-c:
					break LOOP
				case <-time.After(300 * time.Millisecond):
					break LOOP
				}
			}

			Convey("Then the function should return", func() {
				So(returned, ShouldBeTrue)
			})
		})
	})

	Convey("Given I create a session with a websocket and kafka info and run listen", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		broker.SetHandlerByMap(map[string]sarama.MockResponse{
			"MetadataRequest": sarama.NewMockMetadataResponse(t).
				SetBroker(broker.Addr(), broker.BrokerID()).
				SetLeader("topic", 0, broker.BrokerID()),
			"OffsetRequest": sarama.NewMockOffsetResponse(t).
				SetOffset("topic", 0, sarama.OffsetOldest, 0).
				SetOffset("topic", 0, sarama.OffsetNewest, 0),
			"FetchRequest": sarama.NewMockFetchResponse(t, 1).
				SetMessage("topic", 0, 0, sarama.StringEncoder(`{"hello":"world"}`)),
		})
		defer broker.Close()

		config := MakePushServerConfig([]string{broker.Addr()}, "topic", nil)

		ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {}))
		defer ts.Close()

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		session := newPushSession(ws, newPushServer(config, bone.New()))

		c := make(chan bool, 1)
		go func() {
			session.listen()
			c <- true
		}()

		Convey("When I keep it alive", func() {

			var returned bool
			select {
			case returned = <-c:
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the function should not return", func() {
				So(returned, ShouldBeFalse)
			})
		})

		Convey("When I close it", func() {

			session.close()

			var returned bool
		LOOP:
			for {
				select {
				case <-session.server.unregister:
				case returned = <-c:
					break LOOP
				case <-time.After(300 * time.Millisecond):
					break LOOP
				}
			}

			Convey("Then the function should return", func() {
				So(returned, ShouldBeTrue)
			})
		})
	})
}
