package bahamut

import (
	"net/http"
	"testing"

	"github.com/aporeto-inc/elemental/test/model"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTracing_extractClaims(t *testing.T) {

	token := "x.eyJyZWFsbSI6IkNlcnRpZmljYXRlIiwiZGF0YSI6eyJjb21tb25OYW1lIjoiYWRtaW4iLCJvcmdhbml6YXRpb24iOiJzeXN0ZW0iLCJvdTpyb290IjoidHJ1ZSIsInJlYWxtIjoiY2VydGlmaWNhdGUiLCJzZXJpYWxOdW1iZXIiOiIxODY3OTg0MjcyNDEzNDMwODM2NzY2MDU2NTk0NDg1NjUxNTk4MTcifSwiYXVkIjoiYXBvcmV0by5jb20iLCJleHAiOjE1MDg1MTYxMzEsImlhdCI6MTUwODQyOTczMSwiaXNzIjoibWlkZ2FyZC5hcG9tdXguY29tIiwic3ViIjoiMTg2Nzk4NDI3MjQxMzQzMDgzNjc2NjA1NjU5NDQ4NTY1MTU5ODE3In0.y"
	tokenInavalid := "eyJyZWFsbSI6IkNlcnRpZmljYXRlIiwiZGF0YSI6eyJjb21tb25OYW1lIjoiYWRtaW4iLCJvcmdhbml6YXRpb24iOiJzeXN0ZW0iLCJvdTpyb290IjoidHJ1ZSIsInJlYWxtIjoiY2VydGlmaWNhdGUiLCJzZXJpYWxOdW1iZXIiOiIxODY3OTg0MjcyNDEzNDMwODM2NzY2MDU2NTk0NDg1NjUxNTk4MTcifSwiYXVkIjoiYXBvcmV0by5jb20iLCJleHAiOjE1MDg1MTYxMzEsImlhdCI6MTUwODQyOTczMSwiaXNzIjoibWlkZ2FyZC5hcG9tdXguY29tIiwic3ViIjoiMTg2Nzk4NDI3MjQxMzQzMDgzNjc2NjA1NjU5NDQ4NTY1MTU5ODE3In0.y"

	Convey("Given I have a Request with Password", t, func() {

		req := elemental.NewRequest()
		req.Password = token

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{"realm":"Certificate","data":{"commonName":"admin","organization":"system","ou:root":"true","realm":"certificate","serialNumber":"186798427241343083676605659448565159817"},"aud":"aporeto.com","exp":1508516131,"iat":1508429731,"iss":"midgard.apomux.com","sub":"186798427241343083676605659448565159817"}`)
			})
		})
	})

	Convey("Given create a request from an http request", t, func() {

		req, _ := http.NewRequest(http.MethodGet, "http://server/lists/xx/tasks?p=v", nil)
		req.Header.Add("X-Namespace", "ns")
		req.Header.Add("Authorization", "Bearer "+token)
		r, _ := elemental.NewRequestFromHTTPRequest(req)

		Convey("When I extract the claims", func() {

			claims := extractClaims(r)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{"realm":"Certificate","data":{"commonName":"admin","organization":"system","ou:root":"true","realm":"certificate","serialNumber":"186798427241343083676605659448565159817"},"aud":"aporeto.com","exp":1508516131,"iat":1508429731,"iss":"midgard.apomux.com","sub":"186798427241343083676605659448565159817"}`)
			})
		})
	})

	Convey("Given I have a Request with invalid token in Password", t, func() {

		req := elemental.NewRequest()
		req.Password = tokenInavalid

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{}`)
			})
		})
	})

	Convey("Given I have a Request with almost invalid token in Password", t, func() {

		req := elemental.NewRequest()
		req.Password = "a.b.c"

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{}`)
			})
		})
	})
}

func TestTracing_tracingName(t *testing.T) {

	Convey("Given I have a create request on some identity", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity

		Convey("When I call tracingName for operation create", func() {

			req.Operation = elemental.OperationCreate
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.create.lists")
			})
		})

		Convey("When I call tracingName for operation update", func() {

			req.Operation = elemental.OperationUpdate
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.update.lists")
			})
		})

		Convey("When I call tracingName for operation delete", func() {

			req.Operation = elemental.OperationDelete
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.delete.lists")
			})
		})

		Convey("When I call tracingName for operation info", func() {

			req.Operation = elemental.OperationInfo
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.info.lists")
			})
		})

		Convey("When I call tracingName for operation retrieve", func() {

			req.Operation = elemental.OperationRetrieve
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.retrieve.lists")
			})
		})

		Convey("When I call tracingName for operation retrieve many", func() {

			req.Operation = elemental.OperationRetrieveMany
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.retrieve_many.lists")
			})
		})

		Convey("When I call tracingName for operation patch", func() {

			req.Operation = elemental.OperationPatch
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.patch.lists")
			})
		})

		Convey("When I call tracingName for operation unknown", func() {

			req.Operation = elemental.Operation("nope")
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "Unknown operation: nope")
			})
		})
	})

}
