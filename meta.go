package bahamut

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aporeto-inc/elemental"
)

// A RouteInfo contains basic information about an api route.
type RouteInfo struct {
	URL     string   `json:"url"`
	Verbs   []string `json:"verbs,omitempty"`
	Private bool     `json:"private,omitempty"`
}

func (r RouteInfo) String() string {
	return fmt.Sprintf("%s -> %s ", r.URL, strings.Join(r.Verbs, ", "))
}

type routeBuilder struct {
	verbs   map[string]struct{}
	private bool
}

func buildVersionedRoutes(factories map[int]elemental.IdentifiableFactory, processorFinder processorFinderFunc) map[int][]RouteInfo {

	addRoute := func(routes map[string]routeBuilder, url string, verb string, private bool) {

		rb, ok := routes[url]
		if !ok {
			rb = routeBuilder{
				verbs:   map[string]struct{}{},
				private: private,
			}
			routes[url] = rb
		}
		rb.verbs[verb] = struct{}{}
	}

	versionedRoutes := map[int][]RouteInfo{}

	for version, factory := range factories {

		versionedRoutes[version] = []RouteInfo{}

		routes := map[string]routeBuilder{}

		for identity, relationship := range factory.Relationships() {

			// If we don't have a processor registered for the given model, we skip.
			if _, err := processorFinder(identity); err != nil {
				continue
			}

			if len(relationship.AllowsCreate) > 0 {
				addRoute(routes, fmt.Sprintf("/%s", identity.Category), "POST", identity.Private)
			}

			if len(relationship.AllowsRetrieve) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "GET", identity.Private)
			}

			if len(relationship.AllowsDelete) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "DELETE", identity.Private)
			}

			if len(relationship.AllowsUpdate) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "PUT", identity.Private)
			}

			for parent := range relationship.AllowsRetrieveMany {

				if parent == "root" {
					addRoute(routes, fmt.Sprintf("/%s", identity.Category), "GET", factory.IdentityFromName(parent).Private)
				} else {
					addRoute(routes, fmt.Sprintf("/%s/:id/%s", parent, identity.Category), "GET", identity.Private)
				}
			}

			for parent := range relationship.AllowsCreate {

				if parent == "root" {
					addRoute(routes, fmt.Sprintf("/%s", identity.Category), "POST", identity.Private)
				} else {
					addRoute(routes, fmt.Sprintf("/%s/:id/%s", parent, identity.Category), "POST", identity.Private)
				}
			}
		}

		for url, rb := range routes {
			var flatVerbs []string

			for v := range rb.verbs {
				flatVerbs = append(flatVerbs, v)
			}

			versionedRoutes[version] = append(
				versionedRoutes[version],
				RouteInfo{
					URL:     url,
					Verbs:   flatVerbs,
					Private: rb.private,
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
