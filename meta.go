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
	"sort"
	"strings"

	"go.aporeto.io/elemental"
)

// A RouteInfo contains basic information about an api route.
type RouteInfo struct {
	Identity string   `msgpack:"identity" json:"identity"`
	URL      string   `msgpack:"url" json:"url"`
	Verbs    []string `msgpack:"verbs,omitempty" json:"verbs,omitempty"`
	Private  bool     `msgpack:"private,omitempty" json:"private,omitempty"`
}

func (r RouteInfo) String() string {
	return fmt.Sprintf("%s -> %s", r.URL, strings.Join(r.Verbs, ", "))
}

type routeBuilder struct {
	verbs    map[string]struct{}
	private  bool
	identity elemental.Identity
}

func buildVersionedRoutes(modelManagers map[int]elemental.ModelManager, processorFinder processorFinderFunc) map[int][]RouteInfo {

	addRoute := func(routes map[string]routeBuilder, identity elemental.Identity, url string, verb string, private bool) {

		rb, ok := routes[url]
		if !ok {
			rb = routeBuilder{
				verbs:    map[string]struct{}{},
				private:  private,
				identity: identity,
			}
			routes[url] = rb
		}
		rb.verbs[verb] = struct{}{}
	}

	versionedRoutes := map[int][]RouteInfo{}

	for version, modelManager := range modelManagers {

		versionedRoutes[version] = []RouteInfo{}

		routes := map[string]routeBuilder{}

		for identity, relationship := range modelManager.Relationships() {

			// If we don't have a processor registered for the given model, we skip.
			if _, err := processorFinder(identity); err != nil {
				continue
			}

			if len(relationship.Create) > 0 {
				addRoute(routes, identity, fmt.Sprintf("/%s", identity.Category), "POST", identity.Private)
			}

			if len(relationship.Retrieve) > 0 {
				addRoute(routes, identity, fmt.Sprintf("/%s/:id", identity.Category), "GET", identity.Private)
			}

			if len(relationship.Delete) > 0 {
				addRoute(routes, identity, fmt.Sprintf("/%s/:id", identity.Category), "DELETE", identity.Private)
			}

			if len(relationship.Update) > 0 {
				addRoute(routes, identity, fmt.Sprintf("/%s/:id", identity.Category), "PUT", identity.Private)
			}

			for parent := range relationship.RetrieveMany {

				if parent == "root" {
					addRoute(routes, identity, fmt.Sprintf("/%s", identity.Category), "GET", identity.Private)
				} else {
					addRoute(routes, identity, fmt.Sprintf("/%s/:id/%s", modelManager.IdentityFromName(parent).Category, identity.Category), "GET", identity.Private)
				}
			}

			for parent := range relationship.Create {

				if parent == "root" {
					addRoute(routes, identity, fmt.Sprintf("/%s", identity.Category), "POST", identity.Private)
				} else {
					addRoute(routes, identity, fmt.Sprintf("/%s/:id/%s", modelManager.IdentityFromName(parent).Category, identity.Category), "POST", identity.Private)
				}
			}
		}

		for url, rb := range routes {
			var flatVerbs []string

			for v := range rb.verbs {
				flatVerbs = append(flatVerbs, v)
			}
			sort.Strings(flatVerbs)

			versionedRoutes[version] = append(
				versionedRoutes[version],
				RouteInfo{
					URL:      url,
					Verbs:    flatVerbs,
					Private:  rb.private,
					Identity: rb.identity.Category,
				},
			)
		}
	}

	for _, ri := range versionedRoutes {
		sort.Slice(ri, func(i int, j int) bool {
			return strings.Compare(ri[i].URL, ri[j].URL) == -1
		})
	}

	return versionedRoutes
}
