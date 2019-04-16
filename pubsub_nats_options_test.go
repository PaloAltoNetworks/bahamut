package bahamut

import (
	"context"
	"crypto/tls"
	"testing"

	nats "github.com/nats-io/go-nats"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBahamut_NATSOption(t *testing.T) {

	n := &natsPubSub{}

	Convey("Calling NATSOptCredentials should work", t, func() {
		NATSOptCredentials("user", "pass")(n)
		So(n.username, ShouldEqual, "user")
		So(n.password, ShouldEqual, "pass")
	})

	Convey("Calling NATSOptClusterID should work", t, func() {
		NATSOptClusterID("cid")(n)
		So(n.clusterID, ShouldEqual, "cid")
	})

	Convey("Calling NATSOptClientID should work", t, func() {
		NATSOptClientID("cid")(n)
		So(n.clientID, ShouldEqual, "cid")
	})

	Convey("Calling NATSOptTLS should work", t, func() {
		tlscfg := &tls.Config{}
		NATSOptTLS(tlscfg)(n)
		So(n.tlsConfig, ShouldEqual, tlscfg)
	})
}

func TestBahamut_PubSubNatsOptionsSubscribe(t *testing.T) {

	c := natsSubscribeConfig{}

	Convey("Calling NATSOptSubscribeQueue should work", t, func() {
		NATSOptSubscribeQueue("queueGroup")(&c)
		So(c.queueGroup, ShouldEqual, "queueGroup")
	})

	Convey("Calling NATSOptSubscribeReplyer should work", t, func() {
		r := func(msg *nats.Msg) []byte { return nil }
		NATSOptSubscribeReplyer(r)(&c)
		So(c.replier, ShouldEqual, r)
	})
}

func TestBahamut_PubSubNatsOptionsPublish(t *testing.T) {

	c := natsPublishConfig{}

	Convey("Calling NATSOptPublishReplyValidator should work", t, func() {
		v := func(msg *nats.Msg) error { return nil }
		NATSOptPublishReplyValidator(context.TODO(), v)(&c)
		So(c.ctx, ShouldEqual, context.TODO())
		So(c.replyValidator, ShouldEqual, v)
	})

	Convey("Calling NATSOptPublishRequireAck should work", t, func() {
		NATSOptPublishRequireAck(context.TODO())(&c)
		So(c.ctx, ShouldEqual, context.TODO())
		So(c.replyValidator(&nats.Msg{Data: ackMessage}), ShouldEqual, nil)
		So(c.replyValidator(&nats.Msg{Data: []byte("hello")}).Error(), ShouldEqual, "invalid ack: hello")
	})
}
