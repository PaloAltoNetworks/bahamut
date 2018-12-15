package bahamut

import (
	"fmt"
	"reflect"
	"testing"

	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func Test_buildVersionedRoutes(t *testing.T) {
	type args struct {
		modelManagers   map[int]elemental.ModelManager
		processorFinder processorFinderFunc
	}
	tests := []struct {
		name string
		args args
		want map[int][]RouteInfo
	}{
		{
			"simple",
			args{
				map[int]elemental.ModelManager{0: testmodel.Manager(), 1: testmodel.Manager()},
				func(identity elemental.Identity) (Processor, error) {
					return mockProcessor{}, nil
				},
			},
			map[int][]RouteInfo{
				0: []RouteInfo{
					{
						URL: "/lists",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/lists/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL: "/lists/:id/tasks",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/lists/:id/users",
						Verbs: []string{
							"GET",
						},
					},
					{
						URL: "/tasks",
						Verbs: []string{
							"POST",
						},
					},
					{
						URL: "/tasks/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL: "/users",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/users/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
				},
				1: []RouteInfo{
					{
						URL: "/lists",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/lists/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL: "/lists/:id/tasks",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/lists/:id/users",
						Verbs: []string{
							"GET",
						},
					},
					{
						URL: "/tasks",
						Verbs: []string{
							"POST",
						},
					},
					{
						URL: "/tasks/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL: "/users",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL: "/users/:id",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
				},
			},
		},
		{
			"error retrieving processor",
			args{
				map[int]elemental.ModelManager{0: testmodel.Manager(), 1: testmodel.Manager()},
				func(identity elemental.Identity) (Processor, error) {
					return nil, fmt.Errorf("boom")
				},
			},
			map[int][]RouteInfo{0: []RouteInfo{}, 1: []RouteInfo{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildVersionedRoutes(tt.args.modelManagers, tt.args.processorFinder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildVersionedRoutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteInfo_String(t *testing.T) {
	type fields struct {
		URL     string
		Verbs   []string
		Private bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"simple",
			fields{
				URL:   "http.com",
				Verbs: []string{"POST", "GET"},
			},
			"http.com -> POST, GET",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RouteInfo{
				URL:     tt.fields.URL,
				Verbs:   tt.fields.Verbs,
				Private: tt.fields.Private,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("RouteInfo.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
