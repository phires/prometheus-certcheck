package certificate

import (
	"crypto/tls"
	"crypto/x509"
)

type HostResult struct {
	Host  string
	Certs []*x509.Certificate
}

func GetCertificatesOfHost(host string) (HostResult, error) {
	result := HostResult{
		Host: host,
	}
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return result, err
	}
	defer conn.Close()

	checkedCerts := make(map[string]struct{})
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}

			result.Certs = append(result.Certs, cert)
		}
	}

	return result, nil
}
