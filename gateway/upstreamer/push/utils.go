package push

import (
	"math/rand"
	"regexp"
	"strings"
)

var vregexp = regexp.MustCompile(`^/v/\d+`)

func getTargetIdentity(path string) string {

	parts := strings.Split(
		strings.TrimPrefix(
			vregexp.ReplaceAllString(path, ""),
			"/",
		),
		"/",
	)

	switch len(parts) {

	case 1:
		return parts[0]
	case 2:
		return parts[0]
	default:
		return parts[2]
	}
}

func pick(len int) (int, int) {

	if len < 2 {
		panic("pick: len must be greater than 2")
	}

	idxs := make([]int, len)
	for i := 0; i < len; i++ {
		idxs[i] = i
	}

	rand.Shuffle(len, func(i, j int) { idxs[i], idxs[j] = idxs[j], idxs[i] })

	return idxs[0], idxs[1]
}

func handleAddServicePing(services servicesConfig, sp ping) bool {

	if sp.Status == serviceStatusGoodbye {
		panic("handleAddServicePing received a goodbye service ping")
	}

	srv, ok := services[sp.Name]
	if !ok {
		srv = newService(sp.Name)
		services[sp.Name] = srv
	}

	// In any case we poke the endpoint. This will
	// only do something if the endpoint is already
	// registered.
	defer srv.pokeEndpoint(sp.Endpoint, sp.Load)

	if srv.hasEndpoint(sp.Endpoint) {
		return false
	}

	// We update the info to the latest.
	srv.routes = sp.Routes
	srv.versions = sp.Versions

	// We register the new endpoint.
	srv.registerEndpoint(sp.Endpoint, sp.Load)

	return true
}

func handleRemoveServicePing(services servicesConfig, sp ping) bool {

	if sp.Status == serviceStatusHello {
		panic("handleRemoveServicePing received a hello service ping")
	}

	srv, ok := services[sp.Name]
	if !ok {
		return false
	}

	if !srv.hasEndpoint(sp.Endpoint) {
		return false
	}

	srv.unregisterEndpoint(sp.Endpoint)

	if len(srv.getEndpoints()) > 0 {
		return true
	}

	delete(services, sp.Name)

	return true
}

func resyncRoutes(services servicesConfig, includePrivate bool, events map[string]string) map[string][]*endpointInfo {

	apis := map[string][]*endpointInfo{}

	for serviceName, config := range services {

		for _, routes := range config.routes {
			for _, route := range routes {
				if !route.Private || includePrivate {
					apis[route.Identity] = append([]*endpointInfo{}, config.getEndpoints()...)
				}
			}
		}

		if api, ok := events[serviceName]; ok {
			apis[api] = append([]*endpointInfo{}, config.getEndpoints()...)
		}
	}

	return apis
}
