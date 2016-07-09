// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

// APIServerConfig holds the configuration for the Bahamut API Server
type APIServerConfig struct {
	TLSCertificatePath string
	TLSKeyPath         string
	TLSCAPath          string
	ListenAddress      string
	Routes             []*Route
	enabled            bool
	EnableProfiling    bool
}

// MakeAPIServerConfig returns a new APIServerConfig
func MakeAPIServerConfig(listen string, caPath, certPath, keyPath string, routes []*Route) APIServerConfig {

	return APIServerConfig{
		TLSCertificatePath: certPath,
		TLSKeyPath:         keyPath,
		TLSCAPath:          caPath,
		ListenAddress:      listen,
		Routes:             routes,
		enabled:            true,
	}
}
