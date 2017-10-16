package mtls

import "crypto/x509"

func makeClaims(cert *x509.Certificate) []string {

	claims := []string{
		"@auth:realm=certificate",
		"@auth:mode=internal",
		"@auth:serialnumber=" + cert.SerialNumber.String(),
		"@auth:commonname=" + cert.Subject.CommonName,
	}

	for _, o := range cert.Subject.Organization {
		claims = append(claims, "@auth:organization="+o)
	}

	for _, ou := range cert.Subject.OrganizationalUnit {
		claims = append(claims, "@auth:organizationalunit="+ou)
	}

	return claims
}
