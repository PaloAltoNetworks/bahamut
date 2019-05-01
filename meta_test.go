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
						URL:      "/lists",
						Identity: "lists",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/lists/:id",
						Identity: "lists",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL:      "/lists/:id/tasks",
						Identity: "tasks",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/lists/:id/users",
						Identity: "users",
						Verbs: []string{
							"GET",
						},
					},
					{
						URL:      "/tasks",
						Identity: "tasks",
						Verbs: []string{
							"POST",
						},
					},
					{
						URL:      "/tasks/:id",
						Identity: "tasks",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL:      "/users",
						Identity: "users",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/users/:id",
						Identity: "users",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
				},
				1: []RouteInfo{
					{
						URL:      "/lists",
						Identity: "lists",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/lists/:id",
						Identity: "lists",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL:      "/lists/:id/tasks",
						Identity: "tasks",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/lists/:id/users",
						Identity: "users",
						Verbs: []string{
							"GET",
						},
					},
					{
						URL:      "/tasks",
						Identity: "tasks",
						Verbs: []string{
							"POST",
						},
					},
					{
						URL:      "/tasks/:id",
						Identity: "tasks",
						Verbs: []string{
							"DELETE",
							"GET",
							"PUT",
						},
					},
					{
						URL:      "/users",
						Identity: "users",
						Verbs: []string{
							"GET",
							"POST",
						},
					},
					{
						URL:      "/users/:id",
						Identity: "users",
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
