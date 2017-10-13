package mtls

import (
	"crypto/x509"
	"net/http"

	"github.com/aporeto-inc/elemental"
)

func verifyPeerCertificates(
	certificates []*x509.Certificate,
	o []string,
	ou []string,
	cn []string,
) error {

	standardErr := elemental.NewError("Forbidden", "Your certificate information has been rejected", "bahamut", http.StatusForbidden)

	if len(certificates) == 0 {
		return elemental.NewError("Forbidden", "This API require mutual TLS authentication and you did not provide any certificate", "bahamut", http.StatusForbidden)
	}

	var ok bool
	var computed int

	for _, cert := range certificates {

		// If the certificate is a CA, skip
		if cert.IsCA {
			continue
		}

		// If the certificate is not meant to be used for
		// client auth, skip
		var valid bool
		for _, u := range cert.ExtKeyUsage {
			if u == x509.ExtKeyUsageClientAuth {
				valid = true
				break
			}
		}
		if !valid {
			continue
		}

		computed++

		ok = false
		// Verify organizations
		if len(o) != 0 {
			for _, org := range cert.Subject.Organization {
				if isStringInSlice(org, o) {
					ok = true
					break
				}
			}
			if !ok {
				return standardErr
			}
		}

		ok = false
		// Verify organizational units
		if len(ou) != 0 {
			for _, org := range cert.Subject.OrganizationalUnit {
				if isStringInSlice(org, ou) {
					ok = true
					break
				}
			}
			if !ok {
				return standardErr
			}
		}

		// Verify common name
		if len(cn) != 0 {
			if !isStringInSlice(cert.Subject.CommonName, cn) {
				return standardErr
			}
		}
	}

	// If we ended up skipping all certs, this is not valid.
	if computed == 0 {
		return standardErr
	}

	return nil
}

func isStringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
