package containers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type dockerProxyConn struct {
	client   net.Conn
	upstream net.Conn
}

// startDockerSocketHTTPSProxy exposes the real local Engine HTTP API through
// verified loopback HTTPS. It exercises the HTTPS Engine and Compose TLS path
// without opening a host-wide Docker TCP listener.
func startDockerSocketHTTPSProxy(t *testing.T, socketPath string) (string, string) {
	t.Helper()
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}
	certDirectory := t.TempDir()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	caTemplate := &x509.Certificate{
		SerialNumber:          randomCertificateSerial(t),
		Subject:               pkix.Name{CommonName: "Blora isolated Docker test CA"},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	serverCertPEM, serverKeyPEM := dockerProxyLeaf(t, ca, caKey, now, "localhost", []net.IP{net.ParseIP("127.0.0.1")}, x509.ExtKeyUsageServerAuth)
	clientCertPEM, clientKeyPEM := dockerProxyLeaf(t, ca, caKey, now, "blora-docker-test-client", nil, x509.ExtKeyUsageClientAuth)
	writeDockerProxyFile(t, certDirectory, "ca.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}))
	writeDockerProxyFile(t, certDirectory, "cert.pem", clientCertPEM)
	writeDockerProxyFile(t, certDirectory, "key.pem", clientKeyPEM)
	serverCertificate, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tlsListener := tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{serverCertificate}, MinVersion: tls.VersionTLS12})
	active := make(map[net.Conn]dockerProxyConn)
	var activeMu sync.Mutex
	var handlers sync.WaitGroup
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		for {
			client, acceptErr := tlsListener.Accept()
			if acceptErr != nil {
				return
			}
			activeMu.Lock()
			active[client] = dockerProxyConn{client: client}
			activeMu.Unlock()
			handlers.Add(1)
			go func(client net.Conn) {
				defer handlers.Done()
				defer func() {
					activeMu.Lock()
					delete(active, client)
					activeMu.Unlock()
					_ = client.Close()
				}()
				upstream, dialErr := net.DialTimeout("unix", socketPath, 5*time.Second)
				if dialErr != nil {
					return
				}
				activeMu.Lock()
				active[client] = dockerProxyConn{client: client, upstream: upstream}
				activeMu.Unlock()
				defer upstream.Close()
				copied := make(chan struct{}, 2)
				go func() {
					_, _ = io.Copy(upstream, client)
					if unix, ok := upstream.(*net.UnixConn); ok {
						_ = unix.CloseWrite()
					}
					copied <- struct{}{}
				}()
				go func() {
					_, _ = io.Copy(client, upstream)
					copied <- struct{}{}
				}()
				<-copied
				_ = client.Close()
				_ = upstream.Close()
				<-copied
			}(client)
		}
	}()
	t.Cleanup(func() {
		_ = tlsListener.Close()
		<-serveDone
		activeMu.Lock()
		for _, pair := range active {
			_ = pair.client.Close()
			if pair.upstream != nil {
				_ = pair.upstream.Close()
			}
		}
		activeMu.Unlock()
		handlers.Wait()
	})
	return "https://" + listener.Addr().String(), certDirectory
}

func dockerProxyLeaf(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey, now time.Time, commonName string, ips []net.IP, usage x509.ExtKeyUsage) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: randomCertificateSerial(t),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
		IPAddresses:  ips,
	}
	if commonName == "localhost" {
		template.DNSNames = []string{"localhost"}
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
}

func randomCertificateSerial(t *testing.T) *big.Int {
	t.Helper()
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal(err)
	}
	return serial
}

func writeDockerProxyFile(t *testing.T, directory, name string, body []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), body, 0600); err != nil {
		t.Fatal(err)
	}
}
