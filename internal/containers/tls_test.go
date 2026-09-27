package containers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEngineMutualTLSAndUntrustedServerRejected(t *testing.T) {
	now := time.Now()
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Blora ephemeral Docker test CA"}, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, _ := x509.ParseCertificate(caDER)
	issue := func(serial int64, usage x509.ExtKeyUsage) ([]byte, []byte, tls.Certificate) {
		key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		template := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "ephemeral Docker endpoint"}, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour), ExtKeyUsage: []x509.ExtKeyUsage{usage}, KeyUsage: x509.KeyUsageDigitalSignature, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}}
		der, e := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
		if e != nil {
			t.Fatal(e)
		}
		pk, e := x509.MarshalPKCS8PrivateKey(key)
		if e != nil {
			t.Fatal(e)
		}
		certPEM, keyPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pk})
		pair, e := tls.X509KeyPair(certPEM, keyPEM)
		if e != nil {
			t.Fatal(e)
		}
		return certPEM, keyPEM, pair
	}
	_, _, serverCert := issue(2, x509.ExtKeyUsageServerAuth)
	clientCert, clientKey, _ := issue(3, x509.ExtKeyUsageClientAuth)
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 {
			t.Error("request lacks authenticated client certificate")
		}
		json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
	}))
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{serverCert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool}
	server.StartTLS()
	defer server.Close()
	dir := t.TempDir()
	for name, data := range map[string][]byte{"ca.pem": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), "cert.pem": clientCert, "key.pem": clientKey} {
		if err = os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	e, err := newEngine(server.URL, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer e.client.CloseIdleConnections()
	if err = e.check(context.Background()); err != nil {
		t.Fatal("mutual TLS failed", err)
	}
	noClient, err := newEngine(server.URL, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer noClient.client.CloseIdleConnections()
	noClient.client.Transport.(*http.Transport).TLSClientConfig.Certificates = nil
	if err = noClient.check(context.Background()); err == nil {
		t.Fatal("mTLS endpoint accepted a client without its certificate")
	}
	wrongName, err := newEngine(server.URL, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer wrongName.client.CloseIdleConnections()
	wrongName.client.Transport.(*http.Transport).TLSClientConfig.ServerName = "wrong.invalid"
	if err = wrongName.check(context.Background()); err == nil {
		t.Fatal("server hostname mismatch was accepted")
	}
	untrusted, err := newEngine(server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	defer untrusted.client.CloseIdleConnections()
	if err = untrusted.check(context.Background()); err == nil {
		t.Fatal("untrusted private CA accepted with system roots")
	}
	if err = os.WriteFile(filepath.Join(dir, "ca.pem"), []byte("invalid CA"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = newEngine(server.URL, dir); err == nil {
		t.Fatal("invalid CA silently ignored")
	}
}
