// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import "net/http"

// APIServerConfig holds the configuration for the Bahamut API Server
type APIServerConfig struct {
	EnableProfiling     bool
	HealthEndpoint      string
	HealthHandler       http.HandlerFunc
	HealthListenAddress string
	ListenAddress       string
	Routes              []*Route
	TLSCAPath           string
	TLSCertificatePath  string
	TLSKeyPath          string

	enabled bool
}

// MakeAPIServerConfig returns a new APIServerConfig
func MakeAPIServerConfig(
	listen string,

	caPath string,
	certPath string,
	keyPath string,

	routes []*Route,

	healthHandler http.HandlerFunc,
	healthListenAddress string,
	healthEndpoint string,
) APIServerConfig {

	return APIServerConfig{
		HealthEndpoint:      healthEndpoint,
		HealthHandler:       healthHandler,
		HealthListenAddress: healthListenAddress,
		ListenAddress:       listen,
		Routes:              routes,
		TLSCAPath:           caPath,
		TLSCertificatePath:  certPath,
		TLSKeyPath:          keyPath,

		enabled: true,
	}
}
