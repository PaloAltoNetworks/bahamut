package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"go.aporeto.io/elemental"
	"go.aporeto.io/elemental/test/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBahamut_Options(t *testing.T) {

	c := config{}

	Convey("Calling OptDisablePanicRecovery should work", t, func() {
		OptDisablePanicRecovery()(&c)
		So(c.general.panicRecoveryDisabled, ShouldEqual, true)
	})

	Convey("Calling OptRestServer should work", t, func() {
		OptRestServer("1.2.3.4:123")(&c)
		So(c.restServer.enabled, ShouldEqual, true)
		So(c.restServer.listenAddress, ShouldEqual, "1.2.3.4:123")
	})

	Convey("Calling OptTimeouts should work", t, func() {
		OptTimeouts(1*time.Second, 2*time.Second, 3*time.Second)(&c)
		So(c.restServer.readTimeout, ShouldEqual, 1*time.Second)
		So(c.restServer.writeTimeout, ShouldEqual, 2*time.Second)
		So(c.restServer.idleTimeout, ShouldEqual, 3*time.Second)
	})

	Convey("Calling OptDisableKeepAlive should work", t, func() {
		OptDisableKeepAlive()(&c)
		So(c.restServer.disableKeepalive, ShouldEqual, true)
	})

	Convey("Calling OptCustomRootHandler should work", t, func() {
		h := func(http.ResponseWriter, *http.Request) {}
		OptCustomRootHandler(h)(&c)
		So(c.restServer.customRootHandlerFunc, ShouldEqual, h)
	})

	Convey("Calling OptPushServer should work", t, func() {
		srv := NewLocalPubSubClient(nil)
		t := "topic"
		OptPushServer(srv, t)(&c)
		So(c.pushServer.enabled, ShouldEqual, true)
		So(c.pushServer.service, ShouldEqual, srv)
		So(c.pushServer.topic, ShouldEqual, t)
	})

	Convey("Calling OptPushDispatchHandler should work", t, func() {
		h := &mockSessionHandler{}
		OptPushDispatchHandler(h)(&c)
		So(c.pushServer.dispatchEnabled, ShouldEqual, true)
		So(c.pushServer.dispatchHandler, ShouldEqual, h)
	})

	Convey("Calling OptPushPublishHandler should work", t, func() {
		h := &mockSessionHandler{}
		OptPushPublishHandler(h)(&c)
		So(c.pushServer.publishEnabled, ShouldEqual, true)
		So(c.pushServer.publishHandler, ShouldEqual, h)
	})

	Convey("Calling OptHealthServer should work", t, func() {
		h := func() error { return nil }
		OptHealthServer("1.2.3.4:123", h)(&c)
		So(c.healthServer.enabled, ShouldEqual, true)
		So(c.healthServer.listenAddress, ShouldEqual, "1.2.3.4:123")
		So(c.healthServer.healthHandler, ShouldEqual, h)
	})

	Convey("Calling OptHealthServerTimeouts should work", t, func() {
		OptHealthServerTimeouts(1*time.Second, 2*time.Second, 3*time.Second)(&c)
		So(c.healthServer.readTimeout, ShouldEqual, 1*time.Second)
		So(c.healthServer.writeTimeout, ShouldEqual, 2*time.Second)
		So(c.healthServer.idleTimeout, ShouldEqual, 3*time.Second)
	})

	Convey("Calling OptProfilingLocal should work", t, func() {
		OptProfilingLocal("1.2.3.4:123")(&c)
		So(c.profilingServer.enabled, ShouldEqual, true)
		So(c.profilingServer.mode, ShouldEqual, "gops")
		So(c.profilingServer.listenAddress, ShouldEqual, "1.2.3.4:123")
	})

	Convey("Calling OptProfilingGCP should work", t, func() {
		OptProfilingGCP("id", "prfx")(&c)
		So(c.profilingServer.enabled, ShouldEqual, true)
		So(c.profilingServer.mode, ShouldEqual, "gcp")
		So(c.profilingServer.gcpProjectID, ShouldEqual, "id")
		So(c.profilingServer.gcpServicePrefix, ShouldEqual, "prfx")
	})

	Convey("Calling OptTLS should work", t, func() {
		certs := []tls.Certificate{}
		r := func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return nil, nil }
		OptTLS(certs, r)(&c)
		So(c.tls.serverCertificates, ShouldResemble, certs)
		So(c.tls.serverCertificatesRetrieverFunc, ShouldEqual, r)
	})

	Convey("Calling OptMTLS should work", t, func() {
		pool := x509.NewCertPool()
		authType := tls.RequestClientCert
		OptMTLS(pool, authType)(&c)
		So(c.tls.clientCAPool, ShouldEqual, pool)
		So(c.tls.authType, ShouldEqual, authType)
	})

	Convey("Calling OptAuthenticators should work", t, func() {
		ra := []RequestAuthenticator{&mockAuth{}}
		rs := []SessionAuthenticator{&mockSessionAuthenticator{}}
		OptAuthenticators(ra, rs)(&c)
		So(c.security.requestAuthenticators, ShouldResemble, ra)
		So(c.security.sessionAuthenticators, ShouldResemble, rs)
	})

	Convey("Calling OptAuthorizers should work", t, func() {
		ra := []Authorizer{&mockAuth{}}
		OptAuthorizers(ra)(&c)
		So(c.security.authorizers, ShouldResemble, ra)
	})

	Convey("Calling OptAuditer should work", t, func() {
		a := &mockAuditer{}
		OptAuditer(a)(&c)
		So(c.security.auditer, ShouldEqual, a)
	})

	Convey("Calling OptRateLimiting should work", t, func() {
		rlm := NewRateLimiter(1)
		OptRateLimiting(rlm)(&c)
		So(c.rateLimiting.rateLimiter, ShouldEqual, rlm)
	})

	Convey("Calling OptModel should work", t, func() {
		m := map[int]elemental.ModelManager{0: testmodel.Manager()}
		OptModel(m)(&c)
		So(c.model.modelManagers, ShouldEqual, m)
	})

	Convey("Calling OptReadOnly should work", t, func() {
		e := []elemental.Identity{testmodel.ListIdentity}
		OptReadOnly(e)(&c)
		So(c.model.readOnly, ShouldEqual, true)
		So(c.model.readOnlyExcludedIdentities, ShouldResemble, e)
	})

	Convey("Calling OptUnmarshallers should work", t, func() {
		u := map[elemental.Identity]CustomUmarshaller{testmodel.ListIdentity: func(*elemental.Request) (elemental.Identifiable, error) { return nil, nil }}
		OptUnmarshallers(u)(&c)
		So(c.model.unmarshallers, ShouldResemble, u)
	})

	Convey("Calling OptMockServer should work", t, func() {
		OptMockServer("a")(&c)
		So(c.mockServer.enabled, ShouldEqual, true)
		So(c.mockServer.listenAddress, ShouldEqual, "a")
	})

	Convey("Calling OptMockServerTimeouts should work", t, func() {
		OptMockServerTimeouts(1*time.Second, 2*time.Second, 3*time.Second)(&c)
		So(c.mockServer.readTimeout, ShouldEqual, 1*time.Second)
		So(c.mockServer.writeTimeout, ShouldEqual, 2*time.Second)
		So(c.mockServer.idleTimeout, ShouldEqual, 3*time.Second)
	})

	Convey("Calling OptServiceInfo should work", t, func() {
		sb := map[string]interface{}{}
		OptServiceInfo("n", "v", sb)(&c)
		So(c.meta.serviceName, ShouldEqual, "n")
		So(c.meta.serviceVersion, ShouldEqual, "v")
		So(c.meta.version, ShouldEqual, sb)
	})

	Convey("Calling OptDisableMetaRoutes should work", t, func() {
		OptDisableMetaRoutes()(&c)
		So(c.meta.disableMetaRoute, ShouldEqual, true)
	})
}
