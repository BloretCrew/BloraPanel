// Package dockerapi holds shared Docker transport configuration. It does not
// grant resource permissions or provide a second container-management stack.
package dockerapi

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TLSConfig uses the system roots by default. An administrator's Docker TLS
// directory adds ca.pem and the client cert.pem/key.pem pair, matching the
// Compose CLI's DOCKER_CERT_PATH. Certificate/hostname checks remain enabled.
func TLSConfig(directory string) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	if directory == "" {
		return config, nil
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	ca, err := os.ReadFile(filepath.Join(directory, "ca.pem"))
	if err != nil {
		return nil, fmt.Errorf("Docker TLS CA file: %w", err)
	}
	if !roots.AppendCertsFromPEM(ca) {
		return nil, errors.New("Docker TLS CA file contains no valid certificates")
	}
	client, err := tls.LoadX509KeyPair(filepath.Join(directory, "cert.pem"), filepath.Join(directory, "key.pem"))
	if err != nil {
		return nil, fmt.Errorf("Docker TLS client certificate/key: %w", err)
	}
	config.RootCAs = roots
	config.Certificates = []tls.Certificate{client}
	return config, nil
}
