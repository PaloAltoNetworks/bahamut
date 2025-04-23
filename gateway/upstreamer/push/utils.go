package push

import (
	"regexp"
	"strings"
)

var vregexp = regexp.MustCompile(`/v/\d+`)

func getTargetIdentity(path string) (string, string) {

	parts := strings.Split(
		strings.TrimPrefix(
			vregexp.ReplaceAllString(path, ""),
			"/",
		),
		"/",
	)

	prefix := ""
	if len(parts) > 1 && parts[0][0] == '_' {
		prefix = parts[0][1:]
		parts = append([]string{}, parts[1:]...)
	}

	switch len(parts) {

	case 1:
		return parts[0], prefix
	case 2:
		return parts[0], prefix
	default:
		return parts[2], prefix
	}
}

func pick(randomizer Randomizer, length int) (int, int) {

	if length < 2 {
		panic("pick: len must be greater than 2")
	}

	idxs := make([]int, len)
	for i := 0; i < length; i++ {
		idxs[i] = i
	}

	randomizer.Shuffle(length, func(i, j int) { idxs[i], idxs[j] = idxs[j], idxs[i] })

	return idxs[0], idxs[1]
}

func handleAddServicePing(services servicesConfig, sp servicePing) bool {

	if sp.Status == entityStatusGoodbye {
		panic("handleAddServicePing received a goodbye service ping")
	}

	srv, ok := services[sp.Key()]
	if !ok {
		srv = newService(sp.Key())
		services[sp.Key()] = srv
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
	srv.registerEndpoint(sp.Endpoint, sp.Load, sp.APILimiters)

	return true
}

func handleRemoveServicePing(services servicesConfig, sp servicePing) bool {

	if sp.Status == entityStatusHello {
		panic("handleRemoveServicePing received a hello service ping")
	}

	srv, ok := services[sp.Key()]
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

	delete(services, sp.Key())

	return true
}

func resyncRoutes(services servicesConfig, includePrivate bool, events map[string]string) map[string][]*endpointInfo {

	apis := map[string][]*endpointInfo{}

	for serviceName, config := range services {

		name, prefix := extractPrefix(serviceName)

		for _, routes := range config.routes {
			for _, route := range routes {
				if !route.Private || includePrivate {
					apis[prefix+"/"+route.Identity] = append([]*endpointInfo{}, config.getEndpoints()...)
				}
			}
		}

		if api, ok := events[name]; ok {
			apis[prefix+"/"+api] = append([]*endpointInfo{}, config.getEndpoints()...)
		}
	}

	return apis
}

func extractPrefix(key string) (name string, prefix string) {

	name = key

	if parts := strings.SplitN(key, "/", 2); len(parts) == 2 {
		prefix = parts[0]
		name = parts[1]
	}

	return name, prefix
}
