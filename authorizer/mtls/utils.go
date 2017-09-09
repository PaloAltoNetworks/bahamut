package mtlsauthorizer

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

	for _, cert := range certificates {

		// Verify organizations
		if len(o) != 0 {
			for _, org := range cert.Subject.Organization {
				if !isStringInSlice(org, o) {
					return standardErr
				}
			}
		}

		// Verify organizational units
		if len(ou) != 0 {
			for _, org := range cert.Subject.OrganizationalUnit {
				if !isStringInSlice(org, ou) {
					return standardErr
				}
			}
		}

		// Verify common name
		if len(cn) != 0 {
			if !isStringInSlice(cert.Subject.CommonName, cn) {
				return standardErr
			}
		}
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
