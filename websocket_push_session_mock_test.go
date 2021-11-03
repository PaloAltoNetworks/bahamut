package bahamut

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
)

func TestMockSession(t *testing.T) {

	Convey("MockSession should work", t, func() {

		s := NewMockSession()
		So(s.MockClaimsMap, ShouldNotBeNil)
		So(s.MockCookies, ShouldNotBeNil)
		So(s.MockHeaders, ShouldNotBeNil)
		So(s.MockParameters, ShouldNotBeNil)

		s.MockClaimsMap = map[string]string{"k": "v"}
		s.MockClientIP = "1.1.1.1"
		s.MockCookies = map[string]*http.Cookie{"c": {}}
		s.MockHeaders = map[string]string{"k": "v"}
		s.MockIdentifier = "id"
		s.MockParameters = map[string]string{"k": "v"}
		s.MockPushConfig = &elemental.PushConfig{}
		s.MockTLSConnectionState = &tls.ConnectionState{}
		s.MockToken = "token"

		s.SetClaims([]string{"k=v"})
		s.SetMetadata("mischief") // A beer to the one who gets the reference.

		So(s.Identifier(), ShouldEqual, "id")
		So(s.Parameter("k"), ShouldEqual, "v")
		So(s.Header("k"), ShouldEqual, "v")
		So(s.PushConfig(), ShouldNotBeNil)
		So(s.Claims(), ShouldResemble, []string{"k=v"})
		So(s.ClaimsMap(), ShouldResemble, map[string]string{"k": "v"})
		So(s.Token(), ShouldEqual, "token")
		So(s.TLSConnectionState(), ShouldNotBeNil)
		So(s.Metadata(), ShouldEqual, "mischief")
		So(s.Context(), ShouldEqual, context.Background())
		So(s.ClientIP(), ShouldEqual, "1.1.1.1")

		cc, err := s.Cookie("c")
		So(cc, ShouldNotBeNil)
		So(err, ShouldBeNil)
		cc, err = s.Cookie("d")
		So(cc, ShouldBeNil)
		So(err, ShouldEqual, http.ErrNoCookie)
	})
}
