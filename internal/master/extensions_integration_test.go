package master

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
)

func TestIndependentReferencePackageResourceBridge(t *testing.T) {
	f := newFileFixture(t)
	var err error
	f.app.extensions, err = extensions.New(t.TempDir(), []string{"window.open", "window.move", "window.close", "shortcut.create", "data.read", "data.write", "resource.read", "resource.write", "notification.publish", "task.create"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../sdk/examples/reference-app/reference.blora-extension.json")
	if err != nil {
		t.Fatal(err)
	}
	var pkg extensions.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatal(err)
	}
	f.admin.request("POST", "/extensions/install-package", json.RawMessage(data), "", 400)
	f.admin.request("POST", "/extensions/install-package", json.RawMessage(data), model.ID(), 201)
	base := "/extensions/" + pkg.Manifest.AppID
	ref := instanceRef(f.instances[0])
	url := base + "/resource?kind=instance&id=" + ref.ID + "&nodeId=" + ref.NodeID
	f.reader.request("GET", url, nil, "", 403)
	grant := model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: pkg.Manifest.AppID}, Action: "app.use"}
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	got := f.reader.request("GET", url, nil, "", 200)
	if string(got["name"]) != `"`+f.instances[0].Name+`"` || got["config"] != nil {
		t.Fatalf("unexpected resource summary: %v", got)
	}
	f.reader.request("GET", base+"/resource?kind=instance&id="+ref.ID+"&nodeId=wrong-node", nil, "", 404)
	f.reader.request("GET", base+"/resource?kind=node&id="+ref.NodeID, nil, "", 403)
	key := model.ID()
	input := map[string]any{"nodeId": ref.NodeID, "payload": map[string]string{"note": "hello 世界"}}
	task := parseFileTask(t, f.admin.request("POST", base+"/tasks", input, key, 202))
	// Acquisition and compilation each have a 30s budget; execution has 5s.
	// Allow these bounded stages to finish under the race build.
	eventually(t, 75*time.Second, func() bool {
		task = parseFileTask(t, f.admin.request("GET", "/tasks/"+task.ID, nil, "", 200))
		return task.State.Terminal()
	})
	if task.State != model.Succeeded {
		t.Fatalf("extension task: %s %s", task.State, task.Error)
	}
	if !strings.Contains(string(task.Result), `"characters":8`) || !strings.Contains(string(task.Result), `"words":2`) {
		t.Fatalf("WASI computation missing: %s", task.Result)
	}
	var terminalAudits int
	if err := f.store.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM audit WHERE actor_id=? AND request_id=? AND result='SUCCEEDED'", task.ActorID, task.RequestID).Scan(&terminalAudits); err != nil || terminalAudits != 1 {
		t.Fatalf("Master terminal audit count=%d error=%v", terminalAudits, err)
	}
	replay := parseFileTask(t, f.admin.request("POST", base+"/tasks", input, key, 202))
	if replay.ID != task.ID {
		t.Fatal("extension replay created another task")
	}
	var update struct {
		Manifest extensions.Manifest `json:"manifest"`
		Payload  []byte              `json:"payload"`
		SHA256   string              `json:"sha256"`
	}
	if err := json.Unmarshal(data, &update); err != nil {
		t.Fatal(err)
	}
	parts, err := extensions.DecodeBundle(update.Payload)
	if err != nil {
		t.Fatal(err)
	}
	parts.Frontend += "\n// upgraded reference package\n"
	encoded, err := json.Marshal(parts)
	if err != nil {
		t.Fatal(err)
	}
	update.Payload = append([]byte(extensions.BundlePrefix), encoded...)
	update.SHA256 = extensions.ModuleHash(update.Payload)
	major, err := strconv.Atoi(strings.Split(strings.TrimPrefix(update.Manifest.PackageVersion, "v"), ".")[0])
	if err != nil {
		t.Fatal(err)
	}
	update.Manifest.PackageVersion = strconv.Itoa(major+1) + ".0.0"
	f.admin.request("POST", base+"/upgrade", update, "", 400)
	f.admin.request("POST", base+"/upgrade", update, model.ID(), 200)
	afterUpgrade := parseFileTask(t, f.admin.request("POST", base+"/tasks", input, key, 202))
	if afterUpgrade.ID != task.ID {
		t.Fatal("upgrade changed replay identity")
	}
	f.admin.request("POST", base+"/tasks", map[string]any{"nodeId": ref.NodeID, "payload": map[string]string{"note": "changed"}}, key, 409)
	f.admin.request("POST", base+"/enabled", map[string]bool{"enabled": false}, "", 400)
	f.admin.request("POST", base+"/enabled", map[string]bool{"enabled": false}, model.ID(), 200)
	f.admin.request("POST", base+"/rollback", nil, "", 400)
	f.admin.request("DELETE", base, nil, "", 400)
	f.reader.request("GET", url, nil, "", 403)
}

func TestExtensionAdminAPIInstallsAndServesPackage(t *testing.T) {
	f := newFileFixture(t)
	f.app.extensions, _ = extensions.New(t.TempDir(), []string{"window.open", "task.create"})
	payload := []byte("bundle-v1")
	data, err := os.ReadFile("../../sdk/examples/reference-app/reference.blora-extension.json")
	if err != nil {
		t.Fatal(err)
	}
	var packaged struct {
		Payload []byte `json:"payload"`
	}
	if err := json.Unmarshal(data, &packaged); err != nil {
		t.Fatal(err)
	}
	parts, err := extensions.DecodeBundle(packaged.Payload)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(extensions.Bundle{Frontend: string(payload), Backend: parts.Backend})
	if err != nil {
		t.Fatal(err)
	}
	payload = append([]byte(extensions.BundlePrefix), encoded...)
	h := sha256.Sum256(payload)
	body := map[string]any{"manifest": map[string]any{"appId": "example.api", "packageVersion": "1.0.0", "hostApiVersion": 1, "title": "API extension", "capabilities": []string{"window.open", "task.create"}}, "sha256": hex.EncodeToString(h[:]), "payload": payload}
	f.admin.request("POST", "/extensions/install-package", body, model.ID(), 201)
	list := f.admin.request("GET", "/extensions", nil, "", 200)
	if len(list) == 0 {
		t.Fatal("empty extension list")
	}
	// The package endpoint is administrator-only and serves the exact verified bytes.
	got := f.admin.raw("GET", "/extensions/example.api/package", nil, "", 200)
	if string(got) != string(payload) {
		t.Fatal("package bytes changed")
	}
	bundle := f.reader.raw("GET", "/extensions/example.api/bundle", nil, "", 200)
	if string(bundle) != "bundle-v1" {
		t.Fatalf("enabled bundle=%q", bundle)
	}
	f.reader.request("GET", "/extensions", nil, "", 403)
	manifest := f.reader.request("GET", "/extensions/example.api/manifest", nil, "", 200)
	if len(manifest) == 0 {
		t.Fatal("manifest response empty")
	}
	secondPayload := []byte("standalone-package")
	secondHash := sha256.Sum256(secondPayload)
	f.admin.request("POST", "/extensions/install-package", map[string]any{
		"manifest": map[string]any{"appId": "invalid-envelope", "packageVersion": "1.0.0", "hostApiVersion": 1, "title": "Invalid", "capabilities": []string{"window.open"}},
		"sha256":   hex.EncodeToString(secondHash[:]), "payload": secondPayload, "unexpected": true,
	}, model.ID(), 400)
	f.admin.request("POST", "/extensions/install-package", map[string]any{
		"manifest": map[string]any{"appId": "standalone.api", "packageVersion": "1.0.0", "hostApiVersion": 1, "title": "Standalone", "capabilities": []string{"window.open"}},
		"sha256":   hex.EncodeToString(secondHash[:]), "payload": secondPayload,
	}, model.ID(), 201)
	// The installed WASI module executes in the selected daemon's sandbox.
	bridgeTask := parseFileTask(t, f.admin.request("POST", "/extensions/example.api/tasks", map[string]any{
		"nodeId": f.instances[0].NodeID, "payload": map[string]any{"kind": "example", "value": 7},
	}, model.ID(), 202))
	deadline := time.Now().Add(75 * time.Second)
	for time.Now().Before(deadline) {
		var err error
		bridgeTask, err = f.store.Task(context.Background(), bridgeTask.ID)
		if err == nil && bridgeTask.State.Terminal() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if bridgeTask.State != model.Succeeded {
		t.Fatalf("extension task state=%s phase=%s error=%s result=%s", bridgeTask.State, bridgeTask.Phase, bridgeTask.Error, bridgeTask.Result)
	}
	if !strings.Contains(string(bridgeTask.Result), `"extensionId":"example.api"`) {
		t.Fatalf("extension task result missing identity: %s", bridgeTask.Result)
	}
	f.reader.request("POST", "/extensions/example.api/tasks", map[string]any{
		"nodeId": f.instances[0].NodeID, "payload": map[string]any{"kind": "reader-must-be-denied"},
	}, model.ID(), 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: "example.api"}, Action: "app.use"}, model.ID(), 200)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "node", ID: f.instances[0].NodeID, NodeID: f.instances[0].NodeID}, Action: "node.read"}, model.ID(), 200)
	readerTask := parseFileTask(t, f.reader.request("POST", "/extensions/example.api/tasks", map[string]any{
		"nodeId": f.instances[0].NodeID, "payload": map[string]any{"kind": "reader-approved"},
	}, model.ID(), 202))
	deadline = time.Now().Add(75 * time.Second)
	for time.Now().Before(deadline) {
		var err error
		readerTask, err = f.store.Task(context.Background(), readerTask.ID)
		if err == nil && readerTask.State.Terminal() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if readerTask.State != model.Succeeded {
		t.Fatalf("approved reader extension task state=%s phase=%s error=%s", readerTask.State, readerTask.Phase, readerTask.Error)
	}
	f.reader.request("GET", "/tasks/"+readerTask.ID, nil, "", 200)
	scoped := "/extensions/example.api/tasks/"
	f.reader.request("GET", scoped+readerTask.ID+"/status", nil, "", 200)
	f.reader.request("GET", "/extensions/standalone.api/tasks/"+readerTask.ID+"/status", nil, "", 404)
	f.reader.request("POST", "/extensions/standalone.api/tasks/"+readerTask.ID+"/cancel", nil, model.ID(), 404)
	f.reader.request("GET", scoped+bridgeTask.ID+"/status", nil, "", 404)
	cancelling := parseFileTask(t, f.reader.request("POST", "/extensions/example.api/tasks", map[string]any{"nodeId": f.instances[0].NodeID, "payload": map[string]string{"note": "cancel me"}}, model.ID(), 202))
	cancelKey := model.ID()
	firstCancel := parseFileTask(t, f.reader.request("POST", scoped+cancelling.ID+"/cancel", nil, cancelKey, 202))
	replayedCancel := parseFileTask(t, f.reader.request("POST", scoped+cancelling.ID+"/cancel", nil, cancelKey, 202))
	if firstCancel.CancellationRequestID != cancelKey || replayedCancel.CancellationRequestID != cancelKey || firstCancel.ID != replayedCancel.ID {
		t.Fatal("extension cancellation retry lost the original receipt")
	}
	eventually(t, 15*time.Second, func() bool {
		cancelling = parseFileTask(t, f.reader.request("GET", scoped+cancelling.ID+"/status", nil, "", 200))
		return cancelling.State.Terminal()
	})
	if cancelling.State != model.Cancelled {
		t.Fatalf("extension cancel state=%s error=%s", cancelling.State, cancelling.Error)
	}
	terminalReplay := parseFileTask(t, f.reader.request("POST", scoped+cancelling.ID+"/cancel", nil, cancelKey, 202))
	if terminalReplay.Revision != cancelling.Revision || terminalReplay.CancellationRequestID != cancelKey {
		t.Fatal("terminal cancellation replay changed the receipt or revision")
	}
	f.admin.request("POST", "/extensions/example.api/enabled", map[string]bool{"enabled": false}, model.ID(), 200)
	f.reader.request("GET", scoped+readerTask.ID+"/status", nil, "", 200)
	f.reader.request("POST", scoped+cancelling.ID+"/cancel", nil, model.ID(), 202)
	f.admin.request("POST", "/extensions/example.api/enabled", map[string]bool{"enabled": true}, model.ID(), 200)
	listed := f.reader.request("GET", "/tasks", nil, "", 200)
	if !strings.Contains(string(listed["items"]), readerTask.ID) {
		t.Fatal("authorized extension task missing from task center")
	}
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: "example.api"}, Action: "app.use"}, model.ID(), 200)
	f.reader.request("GET", "/tasks/"+readerTask.ID, nil, "", 404)
	f.reader.request("GET", scoped+readerTask.ID+"/status", nil, "", 404)
	f.reader.request("POST", scoped+cancelling.ID+"/cancel", nil, model.ID(), 404)
	f.admin.request("POST", "/extensions/example.api/tasks", map[string]any{
		"nodeId": f.instances[0].NodeID, "payload": strings.Repeat("x", 64<<10+1),
	}, model.ID(), 413)
	f.admin.request("POST", "/extensions/example.api/enabled", map[string]any{"enabled": false}, model.ID(), 200)
	f.reader.request("GET", "/extensions/example.api/bundle", nil, "", 409)
	f.admin.request("POST", "/extensions/example.api/tasks", map[string]any{
		"nodeId": f.instances[0].NodeID, "payload": map[string]any{"kind": "must-be-denied"},
	}, model.ID(), 409)
}

func TestExtensionCatalogBrowseAcquireAndUpgrade(t *testing.T) {
	f := newFileFixture(t)
	installRoot := t.TempDir()
	catalogRoot := t.TempDir()
	var err error
	f.app.extensions, err = extensions.New(installRoot, []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := extensions.NewCatalog(catalogRoot)
	if err != nil {
		t.Fatal(err)
	}
	f.app.extensionCatalog = catalog
	makePackage := func(version string, body string) extensions.Package {
		payload := []byte(body)
		h := sha256.Sum256(payload)
		return extensions.Package{Manifest: extensions.Manifest{AppID: "catalog.app", PackageVersion: version, HostAPIVersion: 1, Title: "Catalog", Capabilities: []string{"window.open"}}, SHA256: hex.EncodeToString(h[:]), Payload: payload}
	}
	if err := catalog.Publish(makePackage("1.0.0", "catalog-v1")); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Publish(makePackage("2.0.0", "catalog-v2")); err != nil {
		t.Fatal(err)
	}
	list := f.admin.request("GET", "/extensions/catalog", nil, "", 200)
	var catalogItems []extensions.CatalogEntry
	if err := json.Unmarshal(list["items"], &catalogItems); err != nil {
		t.Fatal(err)
	}
	if len(catalogItems) != 2 {
		t.Fatalf("catalog list length=%d", len(catalogItems))
	}
	f.reader.request("GET", "/extensions/catalog", nil, "", 403)
	got := f.admin.request("GET", "/extensions/catalog/catalog.app/1.0.0", nil, "", 200)
	if len(got) == 0 {
		t.Fatal("catalog acquisition response empty")
	}
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": "catalog.app", "version": "1.0.0"}, "", 400)
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": "catalog.app", "version": "1.0.0"}, model.ID(), 201)
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": "catalog.app", "version": "2.0.0"}, model.ID(), 200)
	installed, err := f.app.extensions.Get("catalog.app")
	if err != nil || installed.Manifest.PackageVersion != "2.0.0" {
		t.Fatalf("catalog upgrade not installed: %+v %v", installed, err)
	}
}

func TestExtensionRemoteCatalogAdminAPI(t *testing.T) {
	f := newFileFixture(t)
	installRoot := t.TempDir()
	if f.app.extensions, _ = extensions.New(installRoot, []string{"window.open"}); f.app.extensions == nil {
		t.Fatal("extension manager unavailable")
	}
	roots, makeServerCertificate := extensionCatalogTestCertificateAuthority(t)
	rotatedRoots, makeRotatedServerCertificate := extensionCatalogTestCertificateAuthority(t)
	firstCertificate := makeServerCertificate(2026092001)
	var activeCertificate atomic.Pointer[tls.Certificate]
	activeCertificate.Store(firstCertificate)
	var servedCertificateSerial atomic.Int64
	type catalogPackage struct {
		manifest extensions.Manifest
		payload  []byte
		sha256   string
	}
	makePackage := func(version, body string) *catalogPackage {
		payload := []byte(body)
		digest := sha256.Sum256(payload)
		return &catalogPackage{
			manifest: extensions.Manifest{AppID: "remote.catalog", PackageVersion: version, HostAPIVersion: 1, Title: "Remote Catalog", Capabilities: []string{"window.open"}},
			payload:  payload,
			sha256:   hex.EncodeToString(digest[:]),
		}
	}
	var activePackage atomic.Pointer[catalogPackage]
	activePackage.Store(makePackage("1.0.0", "remote-catalog-bundle"))
	manifest := activePackage.Load().manifest
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Model a nontrivial registry RTT while keeping the complete Master API
		// flow deterministic and local to this test.
		time.Sleep(80 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		current := activePackage.Load()
		switch r.URL.Path {
		case "/registry/index.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []extensions.CatalogEntry{{Manifest: current.manifest, SHA256: current.sha256, Size: int64(len(current.payload))}}})
		case "/registry/remote.catalog/" + current.manifest.PackageVersion + ".json":
			_ = json.NewEncoder(w).Encode(map[string]any{"manifest": current.manifest, "sha256": current.sha256, "payload": current.payload})
		default:
			http.NotFound(w, r)
		}
	}))
	server.TLS = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{*firstCertificate},
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			certificate := activeCertificate.Load()
			servedCertificateSerial.Store(certificate.Leaf.SerialNumber.Int64())
			return certificate, nil
		},
	}
	server.StartTLS()
	defer server.Close()
	port := server.Listener.Addr().(*net.TCPAddr).Port
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
		Proxy:           nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		},
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	rawURL := "https://" + net.JoinHostPort("catalog.test", strconv.Itoa(port)) + "/registry"
	remote, err := extensions.NewRemoteCatalog(rawURL, client)
	if err != nil {
		t.Fatal(err)
	}
	f.app.extensionCatalog = remote
	listStarted := time.Now()
	list := f.admin.request("GET", "/extensions/catalog", nil, "", 200)
	if elapsed := time.Since(listStarted); elapsed < 70*time.Millisecond {
		t.Fatalf("remote HTTPS list returned before the injected latency: %v", elapsed)
	}
	if servedCertificateSerial.Load() != firstCertificate.Leaf.SerialNumber.Int64() {
		t.Fatalf("catalog browse used TLS certificate serial %d", servedCertificateSerial.Load())
	}
	var items []extensions.CatalogEntry
	if err := json.Unmarshal(list["items"], &items); err != nil || len(items) != 1 {
		t.Fatalf("remote catalog list: %v %+v", err, items)
	}
	f.reader.request("GET", "/extensions/catalog", nil, "", 403)
	secondCertificate := makeServerCertificate(2026092002)
	activeCertificate.Store(secondCertificate)
	// Force a fresh TLS handshake so installation exercises the renewed leaf,
	// not an already authenticated keep-alive connection.
	transport.CloseIdleConnections()
	installStarted := time.Now()
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": manifest.AppID, "version": manifest.PackageVersion}, model.ID(), 201)
	if elapsed := time.Since(installStarted); elapsed < 70*time.Millisecond {
		t.Fatalf("remote HTTPS install returned before the injected latency: %v", elapsed)
	}
	if servedCertificateSerial.Load() != secondCertificate.Leaf.SerialNumber.Int64() {
		t.Fatalf("catalog install did not use renewed TLS certificate serial: %d", servedCertificateSerial.Load())
	}
	installed, err := f.app.extensions.Get(manifest.AppID)
	if err != nil || installed.Manifest.PackageVersion != manifest.PackageVersion {
		t.Fatalf("remote catalog install: %+v %v", installed, err)
	}

	// A new CA root must not be accepted under the old trust configuration.
	// Keep the same hostname and API route, rotate both the root and package,
	// and force the old transport to perform a fresh TLS handshake.
	rotatedManifest := manifest
	rotatedManifest.PackageVersion = "1.1.0"
	rotatedPackage := makePackage(rotatedManifest.PackageVersion, "remote-catalog-bundle-new-root")
	activePackage.Store(rotatedPackage)
	rootCertificate := makeRotatedServerCertificate(2026092101)
	activeCertificate.Store(rootCertificate)
	transport.CloseIdleConnections()
	f.admin.request("GET", "/extensions/catalog", nil, "", 500)
	if servedCertificateSerial.Load() != rootCertificate.Leaf.SerialNumber.Int64() {
		t.Fatalf("old trust did not evaluate the rotated root certificate: %d", servedCertificateSerial.Load())
	}
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": rotatedManifest.AppID, "version": rotatedManifest.PackageVersion}, model.ID(), 404)
	installed, err = f.app.extensions.Get(manifest.AppID)
	if err != nil || installed.Manifest.PackageVersion != manifest.PackageVersion {
		t.Fatalf("untrusted root changed installed extension: %+v %v", installed, err)
	}

	// Explicitly replacing the trusted roots creates a new transport. Only then
	// can the new catalog root be used to browse and install the newer package.
	rotatedTransport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: rotatedRoots},
		Proxy:           nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		},
	}
	defer rotatedTransport.CloseIdleConnections()
	rotatedClient := &http.Client{Transport: rotatedTransport, Timeout: 5 * time.Second}
	rotatedCatalog, err := extensions.NewRemoteCatalog(rawURL, rotatedClient)
	if err != nil {
		t.Fatal(err)
	}
	f.app.extensionCatalog = rotatedCatalog
	list = f.admin.request("GET", "/extensions/catalog", nil, "", 200)
	if err := json.Unmarshal(list["items"], &items); err != nil || len(items) != 1 || items[0].Manifest.PackageVersion != rotatedManifest.PackageVersion {
		t.Fatalf("catalog after explicit root update: %v %+v", err, items)
	}
	if servedCertificateSerial.Load() != rootCertificate.Leaf.SerialNumber.Int64() {
		t.Fatalf("updated trust did not use the rotated root certificate: %d", servedCertificateSerial.Load())
	}
	f.admin.request("POST", "/extensions/catalog/install", map[string]string{"appId": rotatedManifest.AppID, "version": rotatedManifest.PackageVersion}, model.ID(), 200)
	installed, err = f.app.extensions.Get(rotatedManifest.AppID)
	if err != nil || installed.Manifest.PackageVersion != rotatedManifest.PackageVersion {
		t.Fatalf("catalog upgrade after explicit root update: %+v %v", installed, err)
	}
}

func extensionCatalogTestCertificateAuthority(t *testing.T) (*x509.CertPool, func(int64) *tls.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Blora Remote Catalog Test CA"},
		NotBefore:    now.Add(-time.Minute), NotAfter: now.Add(24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	makeLeaf := func(serial int64) *tls.Certificate {
		leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		template := &x509.Certificate{
			SerialNumber: big.NewInt(serial),
			Subject:      pkix.Name{CommonName: "catalog.test"},
			NotBefore:    now.Add(-time.Minute), NotAfter: now.Add(12 * time.Hour),
			KeyUsage:    x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			DNSNames:    []string{"catalog.test"},
		}
		der, err := x509.CreateCertificate(rand.Reader, template, ca, &leafKey.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		leaf, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: leafKey, Leaf: leaf}
	}
	return roots, makeLeaf
}
