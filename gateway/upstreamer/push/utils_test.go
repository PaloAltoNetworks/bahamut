package push

import (
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
)

func Test_getTargetIdentity(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"/",
			args{
				"/",
			},
			"",
		},
		{
			"/users",
			args{
				"/users",
			},
			"users",
		},
		{
			"/users/id",
			args{
				"/users/id",
			},
			"users",
		},
		{
			"/users/id/groups",
			args{
				"/users/id/groups",
			},
			"groups",
		},
		{
			"/v/1/users",
			args{
				"/v/1/users",
			},
			"users",
		},
		{
			"/v/1/users/id",
			args{
				"/v/1/users/id",
			},
			"users",
		},
		{
			"/v/1/users/id/groups",
			args{
				"/v/1/users/id/groups",
			},
			"groups",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTargetIdentity(tt.args.path); got != tt.want {
				t.Errorf("getTargetIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleServicePings(t *testing.T) {

	// TODO: CHECK ROUTES AND VERSIONS

	Convey("Given I have an empty servicesConfig", t, func() {

		scfg := servicesConfig{}

		// Registration

		Convey("When I send a hello ping from an instance of a service", func() {

			routes11 := map[int][]bahamut.RouteInfo{
				0: {
					{
						Identity: "kittens",
						URL:      "/kittens",
						Verbs:    []string{http.MethodDelete},
						Private:  false,
					},
				},
			}

			routes12 := map[int][]bahamut.RouteInfo{
				0: {
					{
						Identity: "cats",
						URL:      "/cats",
						Verbs:    []string{http.MethodGet},
						Private:  false,
					},
				},
			}

			routes2 := map[int][]bahamut.RouteInfo{
				0: {
					{
						Identity: "dogs",
						URL:      "/dogs",
						Verbs:    []string{http.MethodPost},
						Private:  false,
					},
				},
			}

			versions11 := map[string]interface{}{"a": 1}
			versions12 := map[string]interface{}{"a": 2}
			versions2 := map[string]interface{}{"b": 2}

			handled := handleAddServicePing(scfg, ping{
				Name:         "srv1",
				Endpoint:     "1.1.1.1:1",
				PushEndpoint: "push1",
				Status:       serviceStatusHello,
				Load:         0.1,
				Routes:       routes11,
				Versions:     versions11,
			})

			Convey("Then it should have registered a new service config", func() {

				So(handled, ShouldBeTrue)
				So(len(scfg), ShouldEqual, 1)
				So(scfg["srv1"], ShouldNotBeNil)

				srv1 := scfg["srv1"]

				So(srv1.name, ShouldEqual, "srv1")
				So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue)
				So(srv1.routes, ShouldResemble, routes11)
				So(srv1.versions, ShouldResemble, versions11)
				So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
				So(len(srv1.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 1)
			})

			Convey("When I send a second ping from another instance of the same service", func() {

				handled := handleAddServicePing(scfg, ping{
					Name:         "srv1",
					Endpoint:     "1.1.1.1:2",
					PushEndpoint: "push1",
					Status:       serviceStatusHello,
					Load:         0.1,
					Routes:       routes12,
					Versions:     versions12,
				})

				Convey("Then it should have registered a second endpoint in the service config", func() {

					So(handled, ShouldBeTrue)
					So(len(scfg), ShouldEqual, 1)
					So(scfg["srv1"], ShouldNotBeNil)

					srv1 := scfg["srv1"]

					So(srv1.name, ShouldEqual, "srv1")
					So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue)
					So(srv1.hasEndpoint("1.1.1.1:2"), ShouldBeTrue)
					So(srv1.routes, ShouldResemble, routes12)
					So(srv1.versions, ShouldResemble, versions12)
					So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
					So(len(srv1.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 2)
				})

				Convey("When I send a hello ping from an instance of a second service", func() {

					handled := handleAddServicePing(scfg, ping{
						Name:         "srv2",
						Endpoint:     "2.2.2.2:1",
						PushEndpoint: "push2",
						Status:       serviceStatusHello,
						Load:         0.2,
						Routes:       routes2,
						Versions:     versions2,
					})

					Convey("Then it should have registered a new service", func() {

						So(handled, ShouldBeTrue)
						So(len(scfg), ShouldEqual, 2)
						So(scfg["srv1"], ShouldNotBeNil)
						So(scfg["srv2"], ShouldNotBeNil)

						srv1 := scfg["srv1"]
						srv2 := scfg["srv2"]

						So(srv2.name, ShouldEqual, "srv2")
						So(srv2.hasEndpoint("2.2.2.2:1"), ShouldBeTrue)
						So(srv2.routes, ShouldResemble, routes2)
						So(srv2.versions, ShouldResemble, versions2)
						So(len(srv2.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
						So(len(srv2.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 1)

						// quick check on srv1
						So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue)
						So(srv1.hasEndpoint("1.1.1.1:2"), ShouldBeTrue)
						So(srv1.routes, ShouldResemble, routes12)
						So(srv1.versions, ShouldResemble, versions12)
						So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
						So(len(srv1.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 2)
					})

					// Unregistration

					Convey("When I send a goodbye ping from an instance of the first service", func() {

						handled := handleRemoveServicePing(scfg, ping{
							Name:         "srv1",
							Endpoint:     "1.1.1.1:1",
							PushEndpoint: "push2",
							Status:       serviceStatusGoodbye,
						})

						Convey("Then it should have unregistered one endpoint of srv1", func() {

							So(handled, ShouldBeTrue)
							So(len(scfg), ShouldEqual, 2)
							So(scfg["srv1"], ShouldNotBeNil)
							So(scfg["srv2"], ShouldNotBeNil)

							srv1 := scfg["srv1"]
							srv2 := scfg["srv2"]

							So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeFalse)
							So(srv1.hasEndpoint("1.1.1.1:2"), ShouldBeTrue)
							So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
							So(len(srv1.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 1)

							// quick check on srv2
							So(srv2.name, ShouldEqual, "srv2")
							So(srv2.hasEndpoint("2.2.2.2:1"), ShouldBeTrue)
							So(len(srv2.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
							So(len(srv2.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 1)
						})

						Convey("When I finally send another goodbye from the last instance of srv1", func() {

							handled := handleRemoveServicePing(scfg, ping{
								Name:         "srv1",
								Endpoint:     "1.1.1.1:2",
								PushEndpoint: "push2",
								Status:       serviceStatusGoodbye,
							})

							Convey("Then it should have unregistered srv1", func() {
								So(handled, ShouldBeTrue)
								So(len(scfg), ShouldEqual, 1)
								So(scfg["srv1"], ShouldBeNil)
								So(scfg["srv2"], ShouldNotBeNil)

								srv2 := scfg["srv2"]

								// quick check on srv2
								So(srv2.name, ShouldEqual, "srv2")
								So(srv2.hasEndpoint("2.2.2.2:1"), ShouldBeTrue)
								So(len(srv2.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
								So(len(srv2.outdatedEndpoints(time.Now().Add(time.Hour))), ShouldEqual, 1)
							})
						})
					})
				})
			})
		})
	})

	Convey("When I send a hello ping from an known service/endpoint", t, func() {

		scfg := servicesConfig{
			"srv1": &service{
				name: "srv1",
				endpoints: map[string]*endpointInfo{
					"1.1.1.1:1": {
						address:  "1.1.1.1:1",
						lastSeen: time.Now().Add(-2 * time.Hour), // looksy here
						lastLoad: 0.1,
					},
				},
			},
		}

		srv1 := scfg["srv1"]

		So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Hour))), ShouldEqual, 1)

		handled := handleAddServicePing(scfg, ping{
			Name:         "srv1",
			Endpoint:     "1.1.1.1:1",
			PushEndpoint: "push1",
			Status:       serviceStatusHello,
			Load:         0.42,
		})

		Convey("Then it just have poked the outdated timer", func() {

			So(handled, ShouldBeFalse)

			So(srv1.name, ShouldEqual, "srv1")
			So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue)
			So(srv1.getEndpoints()[0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
			So(len(srv1.outdatedEndpoints(time.Now().Add(-time.Second))), ShouldEqual, 0)
		})
	})

	Convey("When I send a goodbye ping from an instance of an unknown service", t, func() {

		now := time.Now()

		scfg := servicesConfig{
			"srv1": &service{
				name: "srv1",
				endpoints: map[string]*endpointInfo{
					"1.1.1.1:1": {
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.1,
					},
				},
			},
		}

		handled := handleRemoveServicePing(scfg, ping{
			Name:         "srv2",
			Endpoint:     "1.1.1.1:1", // looksy here
			PushEndpoint: "push1",
			Status:       serviceStatusGoodbye,
			Load:         0.2,
		})

		Convey("Then it should have ignored the ping", func() {

			So(len(scfg), ShouldEqual, 1)
			So(handled, ShouldBeFalse)

			srv1 := scfg["srv1"]

			So(srv1.name, ShouldEqual, "srv1")
			So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue) // looksy there
		})
	})

	Convey("When I send a goodbye ping from an instance of an endpoint", t, func() {

		now := time.Now()

		scfg := servicesConfig{
			"srv1": &service{
				name: "srv1",
				endpoints: map[string]*endpointInfo{
					"1.1.1.1:1": {
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.1,
					},
				},
			},
		}

		handled := handleRemoveServicePing(scfg, ping{
			Name:         "srv1",
			Endpoint:     "2.2.2.2:1",
			PushEndpoint: "push1",
			Status:       serviceStatusGoodbye,
			Load:         0.2,
		})

		Convey("Then it should have ignored the ping", func() {

			So(len(scfg), ShouldEqual, 1)
			So(handled, ShouldBeFalse)

			srv1 := scfg["srv1"]

			So(srv1.name, ShouldEqual, "srv1")
			So(srv1.hasEndpoint("1.1.1.1:1"), ShouldBeTrue)
			So(srv1.hasEndpoint("2.2.2.2:1"), ShouldBeFalse)
		})
	})

	Convey("Calling handleAddServicePing with a goodbye service ping should panic", t, func() {

		scfg := servicesConfig{}

		So(func() {
			handleAddServicePing(scfg, ping{
				Status: serviceStatusGoodbye,
			})
		}, ShouldPanicWith, "handleAddServicePing received a goodbye service ping")
	})

	Convey("Calling handleRemoveServicePing with a goodbye service ping should panic", t, func() {

		scfg := servicesConfig{}

		So(func() {
			handleRemoveServicePing(scfg, ping{
				Status: serviceStatusHello,
			})
		}, ShouldPanicWith, "handleRemoveServicePing received a hello service ping")
	})
}

func Test_resyncRoutes(t *testing.T) {

	now := time.Now()

	type args struct {
		services       servicesConfig
		includePrivate bool
		events         map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string][]*endpointInfo
	}{
		{
			"simple",
			args{
				servicesConfig{
					"srv1": &service{
						name: "srv1",
						routes: map[int][]bahamut.RouteInfo{
							0: {
								{
									Identity: "cats",
									URL:      "/cats",
									Verbs:    []string{http.MethodGet},
									Private:  false,
								},
								{
									Identity: "kittens",
									URL:      "/kittens",
									Verbs:    []string{http.MethodDelete},
									Private:  true,
								},
							},
						},
						versions: map[string]interface{}{
							"hello": "hey",
						},
						endpoints: map[string]*endpointInfo{
							"1.1.1.1:1": {
								address:  "1.1.1.1:1",
								lastSeen: now,
								lastLoad: 0.0,
							},
							"1.1.1.1:2": {
								address:  "1.1.1.1:2",
								lastSeen: now,
								lastLoad: 0.0,
							},
						},
					},
				},
				true,
				map[string]string{},
			},
			map[string][]*endpointInfo{
				"cats": {
					{
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.0,
					},
					{
						address:  "1.1.1.1:2",
						lastSeen: now,
						lastLoad: 0.0,
					},
				},
				"kittens": {
					{
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.0,
					},
					{
						address:  "1.1.1.1:2",
						lastSeen: now,
						lastLoad: 0.0,
					},
				},
			},
		},

		{
			"without private",
			args{
				servicesConfig{
					"srv1": &service{
						name: "srv1",
						routes: map[int][]bahamut.RouteInfo{
							0: {
								{
									Identity: "cats",
									URL:      "/cats",
									Verbs:    []string{http.MethodGet},
									Private:  false,
								},
								{
									Identity: "kittens",
									URL:      "/kittens",
									Verbs:    []string{http.MethodDelete},
									Private:  true,
								},
							},
						},
						versions: map[string]interface{}{
							"hello": "hey",
						},
						endpoints: map[string]*endpointInfo{
							"1.1.1.1:1": {
								address:  "1.1.1.1:1",
								lastSeen: now,
								lastLoad: 0.0,
							},
							"1.1.1.1:2": {
								address:  "1.1.1.1:2",
								lastSeen: now,
								lastLoad: 0.0,
							},
						},
					},
				},
				false,
				map[string]string{},
			},
			map[string][]*endpointInfo{
				"cats": {
					{
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.0,
					},
					{
						address:  "1.1.1.1:2",
						lastSeen: now,
						lastLoad: 0.0,
					},
				},
			},
		},

		{
			"with events",
			args{
				servicesConfig{
					"srv1": &service{
						name: "srv1",
						routes: map[int][]bahamut.RouteInfo{
							0: {
								{
									Identity: "cats",
									URL:      "/cats",
									Verbs:    []string{http.MethodGet},
									Private:  false,
								},
								{
									Identity: "kittens",
									URL:      "/kittens",
									Verbs:    []string{http.MethodDelete},
									Private:  true,
								},
							},
						},
						versions: map[string]interface{}{
							"hello": "hey",
						},
						endpoints: map[string]*endpointInfo{
							"1.1.1.1:1": {
								address:  "1.1.1.1:1",
								lastSeen: now,
								lastLoad: 0.0,
							},
							"1.1.1.1:2": {
								address:  "1.1.1.1:2",
								lastSeen: now,
								lastLoad: 0.0,
							},
						},
					},
				},
				false,
				map[string]string{"srv1": "evt1"},
			},
			map[string][]*endpointInfo{
				"cats": {
					{
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.0,
					},
					{
						address:  "1.1.1.1:2",
						lastSeen: now,
						lastLoad: 0.0,
					},
				},
				"evt1": {
					{
						address:  "1.1.1.1:1",
						lastSeen: now,
						lastLoad: 0.0,
					},
					{
						address:  "1.1.1.1:2",
						lastSeen: now,
						lastLoad: 0.0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := resyncRoutes(tt.args.services, tt.args.includePrivate, tt.args.events)

			// sort this shit
			for _, v := range tt.want {
				sort.Slice(v, func(i, j int) bool {
					return strings.Compare(v[i].address, v[j].address) == -1
				})
			}

			for _, v := range got {
				sort.Slice(v, func(i, j int) bool {
					return strings.Compare(v[i].address, v[j].address) == -1
				})
			}

			if ShouldResemble(got, tt.want) != "" {
				t.Errorf("resyncRoutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPick(t *testing.T) {

	Convey("calling pick with len lesser than 2 should panic", t, func() {
		So(func() { pick(-1) }, ShouldPanicWith, "pick: len must be greater than 2")
		So(func() { pick(0) }, ShouldPanicWith, "pick: len must be greater than 2")
		So(func() { pick(1) }, ShouldPanicWith, "pick: len must be greater than 2")
	})

	// Since this function is random by nature, these tests
	// ar just ensuring very basic behavior
	Convey("Given have a len of 2", t, func() {

		i1, i2 := pick(2)

		Convey("Then i1 and i2 should be correct", func() {
			So(i1, ShouldBeBetweenOrEqual, 0, 1)
			So(i2, ShouldBeBetweenOrEqual, 0, 1)
			So(i2, ShouldNotEqual, i1)
		})
	})

	Convey("Given have a len of 2", t, func() {

		i1, i2 := pick(10)

		Convey("Then i1 and i2 should be correct", func() {
			So(i1, ShouldBeBetweenOrEqual, 0, 9)
			So(i2, ShouldBeBetweenOrEqual, 0, 9)
			So(i2, ShouldNotEqual, i1)
		})
	})
}
