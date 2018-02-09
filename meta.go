package bahamut

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aporeto-inc/elemental"
)

// A RouteInfo contains basic information about an api route.
type RouteInfo struct {
	URL   string   `json:"url"`
	Verbs []string `json:"verbs,omitempty"`
}

func (r RouteInfo) String() string {
	return fmt.Sprintf("%s -> %s ", r.URL, strings.Join(r.Verbs, ", "))
}

func buildVersionedRoutes(registry map[int]elemental.RelationshipsRegistry) map[int][]RouteInfo {

	addRoute := func(routes map[string]map[string]struct{}, url string, verb string) {

		verbs, ok := routes[url]
		if !ok {
			verbs = map[string]struct{}{}
			routes[url] = verbs
		}
		verbs[verb] = struct{}{}
	}

	versionedRoutes := map[int][]RouteInfo{}

	for version, relationships := range registry {

		versionedRoutes[version] = []RouteInfo{}

		routes := map[string]map[string]struct{}{}

		for identity, relationship := range relationships {

			if len(relationship.AllowsCreate) > 0 {
				addRoute(routes, fmt.Sprintf("/%s", identity.Category), "POST")
			}

			if len(relationship.AllowsRetrieve) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "GET")
			}

			if len(relationship.AllowsDelete) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "DELETE")
			}

			if len(relationship.AllowsUpdate) > 0 {
				addRoute(routes, fmt.Sprintf("/%s/:id", identity.Category), "PUT")
			}

			for parent := range relationship.AllowsRetrieveMany {
				if parent == "root" {
					addRoute(routes, fmt.Sprintf("/%s", identity.Category), "GET")
				} else {
					addRoute(routes, fmt.Sprintf("/%s/:id/%s", parent, identity.Category), "GET")
				}
			}

			for parent := range relationship.AllowsCreate {
				if parent == "root" {
					addRoute(routes, fmt.Sprintf("/%s", identity.Category), "POST")
				} else {
					addRoute(routes, fmt.Sprintf("/%s/:id/%s", parent, identity.Category), "POST")
				}
			}
		}

		for url, verbs := range routes {
			var flatVerbs []string

			for v := range verbs {
				flatVerbs = append(flatVerbs, v)
			}

			versionedRoutes[version] = append(
				versionedRoutes[version],
				RouteInfo{
					URL:   url,
					Verbs: flatVerbs,
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
