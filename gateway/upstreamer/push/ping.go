package push

import (
	"go.aporeto.io/bahamut"
)

type serviceStatus int

const (
	serviceStatusGoodbye serviceStatus = iota
	serviceStatusHello
)

type ping struct {
	// Decodable: must be public
	Name         string
	Endpoint     string
	PushEndpoint string
	Status       serviceStatus
	Routes       map[int][]bahamut.RouteInfo
	Versions     map[string]interface{}
	Load         float64
}
