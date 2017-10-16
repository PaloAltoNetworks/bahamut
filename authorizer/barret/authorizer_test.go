package barret

import (
	"net/http"
	"testing"
	"time"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
	"github.com/aporeto-inc/manipulate"

	"github.com/aporeto-inc/manipulate/maniptest"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBahamut_BarretAuthorizer(t *testing.T) {

	Convey("Given I have barretAuthorizer", t, func() {

		m := maniptest.NewTestManipulator()
		a := NewBarretAuthorizer(m, 1*time.Second)
		ctx := bahamut.NewContext()
		ctx.Request.Password = "atoken"

		Convey("When I call isAuthorized and the cert is not revoked", func() {

			ctx.SetClaims([]string{"@auth:realm=certificate", "@auth:serialnumber=xxxx"})

			var expectedSerialNumber string
			m.MockRetrieve(t, func(ctx *manipulate.Context, objects ...elemental.Identifiable) error {
				expectedSerialNumber = objects[0].Identifier()
				return nil
			})

			action, err := a.IsAuthorized(ctx)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then ok should be true", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})

			Convey("Then expectedSerialNumber should be xxxx", func() {
				So(expectedSerialNumber, ShouldEqual, "xxxx")
			})
		})

		Convey("When I call isAuthorized and the cert is revoked", func() {

			ctx.SetClaims([]string{"@auth:realm=certificate", "@auth:serialnumber=yyyy"})

			var expectedSerialNumber string
			m.MockRetrieve(t, func(ctx *manipulate.Context, objects ...elemental.Identifiable) error {
				expectedSerialNumber = objects[0].Identifier()
				return elemental.NewError("revoked", "completely revoked dude", "test", http.StatusForbidden)
			})

			action, err := a.IsAuthorized(ctx)
			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then ok should be false", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})

			Convey("Then expectedSerialNumber should be yyyy", func() {
				So(expectedSerialNumber, ShouldEqual, "yyyy")
			})
		})

		Convey("When I call isAuthorized and the cert is revoked and I try right away and it's not anymore", func() {

			ctx.SetClaims([]string{"@auth:realm=certificate", "@auth:serialnumber=yyyy"})

			var callN int
			m.MockRetrieve(t, func(ctx *manipulate.Context, objects ...elemental.Identifiable) error {
				callN++
				return elemental.NewError("revoked", "completely revoked dude", "test", http.StatusForbidden)
			})

			action1, err1 := a.IsAuthorized(ctx)
			action2, err2 := a.IsAuthorized(ctx)
			Convey("Then err1 should not be nil", func() {
				So(err1, ShouldNotBeNil)
			})
			Convey("Then err2 should be nil", func() {
				So(err2, ShouldBeNil)
			})

			Convey("Then ok1 should be false", func() {
				So(action1, ShouldEqual, bahamut.AuthActionKO)
			})
			Convey("Then ok2 should be false", func() {
				So(action2, ShouldEqual, bahamut.AuthActionKO)
			})

			Convey("Then callN should be 1", func() {
				So(callN, ShouldEqual, 1)
			})
		})

		Convey("When I call isAuthorized and but the token is not from a cert", func() {

			ctx.SetClaims([]string{"@auth:realm=not-certificate", "@auth:serialnumber=yyyy"})

			var called bool
			m.MockRetrieve(t, func(ctx *manipulate.Context, objects ...elemental.Identifiable) error {
				called = true
				return nil
			})

			action, err := a.IsAuthorized(ctx)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then ok should be true", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})

			Convey("Then called should be falsed", func() {
				So(called, ShouldBeFalse)
			})
		})
	})
}
