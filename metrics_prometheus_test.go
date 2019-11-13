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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_sanitizeURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test /toto",
			args{
				"/toto",
			},
			"/toto",
		},
		{
			"test /v/1/toto",
			args{
				"/v/1/toto",
			},
			"/toto",
		},
		{
			"test /toto/xxxxxxx",
			args{
				"/toto/xxxxxxx",
			},
			"/toto/:id",
		},
		{
			"test /v/1/toto/xxxxxxx",
			args{
				"/v/1/toto/xxxxxxx",
			},
			"/toto/:id",
		},
		{
			"test /toto/xxxxxxx/titi",
			args{
				"/toto/xxxxxxx/titi",
			},
			"/toto/:id/titi",
		},
		{
			"test /v/1/toto/xxxxxxx/titi",
			args{
				"/v/1/toto/xxxxxxx/titi",
			},
			"/toto/:id/titi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeURL(tt.args.url); got != tt.want {
				t.Errorf("sanitizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeasureRequest(t *testing.T) {

	Convey("Given I have a PrometheusMetricsManager", t, func() {

		r := prometheus.NewRegistry()
		pmm := newPrometheusMetricsManager(r).(*prometheusMetricsManager)

		Convey("When I call measure a valid request", func() {

			f := pmm.MeasureRequest("GET", "/toto/id")
			f(200, nil)

			data, _ := r.Gather()

			Convey("Then the data should collected", func() {
				So(data[1].GetName(), ShouldEqual, "http_requests_total")
				So(data[1].GetMetric()[0].Counter.String(), ShouldEqual, "value:1 ")
				So(data[1].GetMetric()[0].Label[0].String(), ShouldEqual, `name:"code" value:"200" `)
				So(data[1].GetMetric()[0].Label[1].String(), ShouldEqual, `name:"method" value:"GET" `)
				So(data[1].GetMetric()[0].Label[2].String(), ShouldEqual, `name:"url" value:"/toto/:id" `)
			})
		})

		Convey("When I call measure a 502 request", func() {

			f := pmm.MeasureRequest("GET", "http://toto.com/id/toto")
			f(502, nil)

			data, _ := r.Gather()

			Convey("Then the data should collected", func() {
				So(data[0].GetName(), ShouldEqual, "http_errors_5xx_total")
				So(data[0].GetMetric()[0].Label[0].String(), ShouldEqual, `name:"code" value:"502" `)
				So(data[0].GetMetric()[0].Label[1].String(), ShouldEqual, `name:"method" value:"GET" `)
				So(data[0].GetMetric()[0].Label[2].String(), ShouldEqual, `name:"trace" value:"unknown" `)
				So(data[0].GetMetric()[0].Label[3].String(), ShouldEqual, `name:"url" value:"http://:id/id/toto" `)
			})
		})
	})
}

func TestRegisterWSConnection(t *testing.T) {

	Convey("Given I have a PrometheusMetricsManager", t, func() {

		r := prometheus.NewRegistry()
		pmm := newPrometheusMetricsManager(r).(*prometheusMetricsManager)

		Convey("When I call RegisterWSConnection twice", func() {

			pmm.RegisterWSConnection()
			pmm.RegisterWSConnection()

			data, _ := r.Gather()

			Convey("Then the total should increase", func() {
				So(data[0].GetName(), ShouldEqual, "http_ws_connections_current")
				So(data[0].GetMetric()[0].String(), ShouldEqual, "gauge:<value:2 > ")
				So(data[1].GetName(), ShouldEqual, "http_ws_connections_total")
				So(data[1].GetMetric()[0].String(), ShouldEqual, "counter:<value:2 > ")
			})

			Convey("When I call UnregisterWSConnection", func() {

				pmm.UnregisterWSConnection()

				data, _ := r.Gather()

				Convey("Then the total should increase", func() {
					So(data[0].GetName(), ShouldEqual, "http_ws_connections_current")
					So(data[0].GetMetric()[0].String(), ShouldEqual, "gauge:<value:1 > ")
					So(data[1].GetName(), ShouldEqual, "http_ws_connections_total")
					So(data[1].GetMetric()[0].String(), ShouldEqual, "counter:<value:2 > ")
				})
			})
		})
	})
}

func TestRegisterTCPConnection(t *testing.T) {

	Convey("Given I have a PrometheusMetricsManager", t, func() {

		r := prometheus.NewRegistry()
		pmm := newPrometheusMetricsManager(r).(*prometheusMetricsManager)

		Convey("When I call RegisterTCPConnection twice", func() {

			pmm.RegisterTCPConnection()
			pmm.RegisterTCPConnection()

			data, _ := r.Gather()

			Convey("Then the total should increase", func() {
				So(data[2].GetName(), ShouldEqual, "tcp_connections_current")
				So(data[2].GetMetric()[0].String(), ShouldEqual, "gauge:<value:2 > ")
				So(data[3].GetName(), ShouldEqual, "tcp_connections_total")
				So(data[3].GetMetric()[0].String(), ShouldEqual, "counter:<value:2 > ")
			})

			Convey("When I call UnregisterTCPConnection", func() {

				pmm.UnregisterTCPConnection()

				data, _ := r.Gather()

				Convey("Then the total should increase", func() {
					So(data[2].GetName(), ShouldEqual, "tcp_connections_current")
					So(data[2].GetMetric()[0].String(), ShouldEqual, "gauge:<value:1 > ")
					So(data[3].GetName(), ShouldEqual, "tcp_connections_total")
					So(data[3].GetMetric()[0].String(), ShouldEqual, "counter:<value:2 > ")
				})
			})
		})
	})
}
