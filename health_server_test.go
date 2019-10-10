// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func freePort() (port int) {
	ln, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		panic(err)
	}
	port = ln.Addr().(*net.TCPAddr).Port
	if err = ln.Close(); err != nil {
		panic(err)
	}
	return
}

// A MetricsManager handles Prometheus Metrics Management
type testMetricsManager struct{}

func (m *testMetricsManager) MeasureRequest(method string, url string) FinishMeasurementFunc {
	return nil
}
func (m *testMetricsManager) RegisterWSConnection()   {}
func (m *testMetricsManager) UnregisterWSConnection() {}
func (m *testMetricsManager) Write(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func TestHealthServer(t *testing.T) {

	Convey("Given I have a health server with no custom handlers", t, func() {

		port := freePort()
		cfg := config{}
		cfg.healthServer.listenAddress = fmt.Sprintf("127.0.0.1:%d", port)

		hs := newHealthServer(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go hs.start(ctx)
		<-time.After(1 * time.Second)

		Convey("When I get /", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 204", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When I get /metrics", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 501", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusNotImplemented)
			})
		})

		Convey("When I get /something with no custom stats", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/something", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 404", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When I send a POST", func() {

			resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/something", port), "", nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 405", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
			})
		})

		Convey("When I stop it", func() {

			hs.stop()

			Convey("Then it should stop", func() {

				resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/something", port), "", nil)

				Convey("Then err should not be nil", func() {
					So(err, ShouldNotBeNil)
				})

				Convey("Then resp should be nil", func() {
					So(resp, ShouldBeNil)
				})
			})
		})
	})
}

func TestHealthServerWithCustomHandler(t *testing.T) {

	Convey("Given I have a health server with custom handlers", t, func() {

		port := freePort()
		cfg := config{}
		cfg.healthServer.listenAddress = fmt.Sprintf("127.0.0.1:%d", port)
		cfg.healthServer.healthHandler = func() error { return fmt.Errorf("boom") }
		cfg.healthServer.metricsManager = &testMetricsManager{}
		cfg.healthServer.customStats = map[string]HealthStatFunc{
			"teapot": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			},
		}

		hs := newHealthServer(cfg)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go hs.start(ctx)
		defer hs.stop()

		<-time.After(1 * time.Second)

		Convey("When I get / with", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 500", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When I get /metrics", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 418", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusTeapot)
			})
		})

		Convey("When I get /teapot", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/teapot", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 418", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusTeapot)
			})
		})

		Convey("When I get /something", func() {

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/something", port))

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then code should be 404", func() {
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
