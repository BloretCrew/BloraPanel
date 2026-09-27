// devfixture starts only task-owned local services and writes fresh credentials
// to a private file for browser integration tests. It is never a production seed.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type account struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
type credentials struct {
	Performance bool     `json:"performance,omitempty"`
	URL         string   `json:"url"`
	Admin       account  `json:"admin"`
	Member      account  `json:"member"`
	NodeIDs     []string `json:"nodeIds"`
	InstanceIDs []string `json:"instanceIds"`
	ExtensionID string   `json:"extensionId,omitempty"`
}
type child struct {
	cmd  *exec.Cmd
	log  *os.File
	done chan struct{}
}
type client struct {
	http      *http.Client
	url, csrf string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	base := flag.String("state-parent", ".local", "parent of a new private fixture directory")
	listen := flag.String("listen", "127.0.0.1:9443", "loopback-only HTTPS address")
	dockerEndpoint := flag.String("docker-endpoint", "", "optional local Docker Engine endpoint for task-owned container tests")
	performance := flag.Bool("performance", false, "seed task-owned 10000-entry directory and 16 MiB transfer file")
	performanceMiB := flag.Int64("performance-transfer-mib", 16, "performance transfer size in MiB (1..1024)")
	flag.Parse()
	if *performanceMiB < 1 || *performanceMiB > 1024 {
		return errors.New("performance-transfer-mib must be 1..1024")
	}
	if *dockerEndpoint != "" && *dockerEndpoint != "unix:///var/run/docker.sock" {
		return errors.New("fixture Docker endpoint must be the local unix socket")
	}
	if *listen != "127.0.0.1:9443" && *listen != "127.0.0.1:9444" {
		return errors.New("fixture supports only loopback ports 9443 or 9444")
	}
	if err := os.MkdirAll(*base, 0700); err != nil {
		return err
	}
	root, err := os.MkdirTemp(*base, "fixture-")
	if err != nil {
		return err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	creds := credentials{URL: "https://" + *listen, Admin: account{Name: "admin", Password: model.ID()}, Member: account{Name: "member", Password: model.ID()}, NodeIDs: []string{}, InstanceIDs: []string{}, ExtensionID: "fixture.reference"}
	creds.Performance = *performance
	masterDir := filepath.Join(root, "master")
	store, err := storage.Open(filepath.Join(masterDir, "master.db"))
	if err != nil {
		return err
	}
	for _, a := range []account{creds.Admin, creds.Member} {
		hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), 12)
		if err != nil {
			store.Close()
			return err
		}
		if _, err := store.CreateUser(context.Background(), model.User{Name: a.Name, Admin: a.Name == creds.Admin.Name}, hash); err != nil {
			store.Close()
			return err
		}
	}
	if err := store.Close(); err != nil {
		return err
	}
	if err := bootstrap.LocalTLS(masterDir); err != nil {
		return err
	}
	extensionsDir := filepath.Join(masterDir, "extensions")
	extensionsCatalogDir := filepath.Join(masterDir, "extensions-catalog")
	if err := seedExtensionCatalog(extensionsCatalogDir); err != nil {
		return err
	}
	pool := x509.NewCertPool()
	certificate, err := os.ReadFile(filepath.Join(masterDir, "tls.crt"))
	if err != nil {
		return err
	}
	pool.AppendCertsFromPEM(certificate)
	jar, _ := cookiejar.New(nil)
	c := &client{http: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}, Jar: jar, Timeout: 5 * time.Second}, url: creds.URL}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var children []*child
	stopChild := func(item *child) {
		if item == nil || item.cmd == nil || item.cmd.Process == nil {
			return
		}
		_ = item.cmd.Process.Signal(os.Interrupt)
		select {
		case <-item.done:
		case <-time.After(10 * time.Second):
			_ = item.cmd.Process.Kill()
			<-item.done
		}
		_ = item.log.Close()
	}
	defer func() {
		c.stopInstances(creds.InstanceIDs)
		for index := len(children) - 1; index >= 0; index-- {
			stopChild(children[index])
		}
	}()
	start := func(binary string, args ...string) error {
		log, err := os.OpenFile(filepath.Join(root, fmt.Sprintf("process-%d.log", len(children))), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		cmd := exec.Command(binary, args...)
		isolateFixtureChild(cmd)
		cmd.Stdout = log
		cmd.Stderr = log
		if err := cmd.Start(); err != nil {
			log.Close()
			return err
		}
		item := &child{cmd: cmd, log: log, done: make(chan struct{})}
		children = append(children, item)
		go func() {
			_ = cmd.Wait()
			close(item.done)
		}()
		return nil
	}
	masterBinary := "./dist/blora-master"
	masterArgs := []string{"--state-dir", masterDir, "--listen", *listen, "--origin", creds.URL, "--static-dir", "web/dist", "--extensions-dir", extensionsDir, "--extensions-catalog", extensionsCatalogDir}
	if err := start(masterBinary, masterArgs...); err != nil {
		return err
	}
	masterChild := children[len(children)-1]
	waitMasterHealth := func() error {
		ready, cancelReady := context.WithTimeout(ctx, 15*time.Second)
		defer cancelReady()
		for {
			req, _ := http.NewRequestWithContext(ready, "GET", creds.URL+"/healthz", nil)
			resp, err := c.http.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
			}
			select {
			case <-ready.Done():
				return errors.New("fixture Master did not become healthy")
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	if err := waitMasterHealth(); err != nil {
		return err
	}
	var loggedIn struct {
		CSRF string `json:"csrfToken"`
	}
	if err := c.call("POST", "/login", map[string]string{"name": creds.Admin.Name, "password": creds.Admin.Password}, &loggedIn); err != nil {
		return err
	}
	c.csrf = loggedIn.CSRF
	for index := 0; index < 2; index++ {
		name := fmt.Sprintf("blora-test-node-%d", index+1)
		var enrolled struct {
			Token string `json:"token"`
		}
		if err := c.call("POST", "/nodes/enrollments", map[string]string{"name": name}, &enrolled); err != nil {
			return err
		}
		tokenFile := filepath.Join(root, name+".enrollment")
		if err := writeJSONOrBytes(tokenFile, []byte(enrolled.Token)); err != nil {
			return err
		}
		config := daemon.Config{StateDir: filepath.Join(root, name), MasterURL: creds.URL, CAFile: filepath.Join(masterDir, "tls.crt"), EnrollmentFile: tokenFile, AllowPGIDFallback: true, DockerEndpoint: *dockerEndpoint}
		configFile := filepath.Join(root, name+".json")
		if err := writeJSONOrBytes(configFile, config); err != nil {
			return err
		}
		if err := start("./dist/blora-daemon", "--config", configFile); err != nil {
			return err
		}
	}
	var nodes struct {
		Items []model.Node `json:"items"`
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		if err := c.call("GET", "/nodes", nil, &nodes); err != nil {
			return err
		}
		online := 0
		for _, n := range nodes.Items {
			if n.State == "ONLINE" {
				online++
			}
		}
		if online == 2 {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("fixture Daemons did not connect; diagnostic files are in %s", root)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	var users struct {
		Items []model.User `json:"items"`
	}
	if err := c.call("GET", "/users", nil, &users); err != nil {
		return err
	}
	var memberID string
	for _, u := range users.Items {
		if u.Name == creds.Member.Name {
			memberID = u.ID
		}
	}
	for index, n := range nodes.Items {
		creds.NodeIDs = append(creds.NodeIDs, n.ID)
		config := model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh", "-c", "printf '中文-BLORA-INSTANCE-LOG\\n'; trap 'exit 0' TERM; while IFS= read -r line; do printf 'CONSOLE:%s\\n' \"$line\"; done"}, StopSeconds: 2, KillSeconds: 2, Escalate: true}
		if *performance {
			config.Directory = filepath.Join(root, fmt.Sprintf("performance-resource-%d", index))
			if err := seedPerformanceResource(config.Directory, index == 0, *performanceMiB<<20); err != nil {
				return err
			}
		}
		var created struct {
			Instance model.Instance `json:"instance"`
		}
		if err := c.call("POST", "/instances", map[string]any{"nodeId": n.ID, "name": fmt.Sprintf("验收实例 %d", index+1), "config": config}, &created); err != nil {
			return err
		}
		creds.InstanceIDs = append(creds.InstanceIDs, created.Instance.ID)
		actions := []string{"instance.read", "file.read"}
		if index == 0 {
			actions = append(actions, "instance.start", "instance.stop", "instance.restart", "file.write")
		}
		for _, action := range actions {
			if err := c.call("POST", "/grants", model.Grant{UserID: memberID, Resource: model.ResourceRef{Kind: "instance", ID: created.Instance.ID, NodeID: n.ID}, Action: action}, nil); err != nil {
				return err
			}
		}
	}
	credentialFile := filepath.Join(root, "browser-credentials.json")
	if err := writeJSONOrBytes(credentialFile, creds); err != nil {
		return err
	}
	fmt.Printf("Fixture ready at %s\nPrivate browser test credentials: %s\nStop with Ctrl+C; only this fixture's instances and processes will be stopped.\n", creds.URL, credentialFile)
	var restartRequests <-chan os.Signal
	if sig := masterRestartSignal(); sig != nil {
		restarts := make(chan os.Signal, 1)
		signal.Notify(restarts, sig)
		defer signal.Stop(restarts)
		restartRequests = restarts
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-restartRequests:
			stopChild(masterChild)
			if err := start(masterBinary, masterArgs...); err != nil {
				return fmt.Errorf("restart fixture Master: %w", err)
			}
			masterChild = children[len(children)-1]
			if err := waitMasterHealth(); err != nil {
				return err
			}
			var loggedIn struct {
				CSRF string `json:"csrfToken"`
			}
			if err := c.call("POST", "/login", map[string]string{"name": creds.Admin.Name, "password": creds.Admin.Password}, &loggedIn); err != nil {
				return fmt.Errorf("re-authenticate fixture client after Master restart: %w", err)
			}
			c.csrf = loggedIn.CSRF
		}
	}
}

func writeJSONOrBytes(path string, value any) error {
	b, ok := value.([]byte)
	if !ok {
		var err error
		b, err = json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
	}
	return os.WriteFile(path, b, 0600)
}
func (c *client) call(method, path string, value, out any) error {
	var body []byte
	if value != nil {
		var err error
		body, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, c.url+"/api/v1"+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", c.csrf)
	req.Header.Set("Idempotency-Key", model.ID())
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fixture %s %s rejected: HTTP %d", method, path, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(out)
	}
	return nil
}
func (c *client) stopInstances(ids []string) {
	// This Master and its database belong exclusively to the fixture. Include
	// instances created through its browser tests, never external nodes.
	var all struct {
		Items []model.Instance `json:"items"`
	}
	if err := c.call("GET", "/instances", nil, &all); err == nil {
		ids = nil
		for _, i := range all.Items {
			ids = append(ids, i.ID)
		}
	}
	for _, id := range ids {
		var result struct {
			Task model.Task `json:"task"`
		}
		if err := c.call("POST", "/instances/"+id+"/actions", map[string]string{"action": "kill"}, &result); err != nil {
			continue
		}
		deadline := time.Now().Add(8 * time.Second)
		for time.Now().Before(deadline) {
			if err := c.call("GET", "/tasks/"+result.Task.ID, nil, &result); err != nil || result.Task.State.Terminal() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// seedExtensionCatalog creates a disposable, task-owned source used by the
// real browser fixture. It deliberately contains two versions so the catalog
// browse, install and upgrade paths are exercised without contacting a network
// registry or embedding a production package in the fixture.
func seedExtensionCatalog(root string) error {
	catalog, err := extensions.NewCatalog(root)
	if err != nil {
		return err
	}
	for _, version := range []string{"1.0.0", "1.1.0"} {
		payload := []byte(fmt.Sprintf("export const start=async call=>{const state=await call('state.capture',null);document.body.dataset.bloraExtension=JSON.stringify({version:%q,state})}", version))
		digest := sha256.Sum256(payload)
		pkg := extensions.Package{
			Manifest: extensions.Manifest{
				AppID: "fixture.reference", PackageVersion: version, HostAPIVersion: 1,
				Title: "Fixture 参考扩展", Icon: "✦", Color: "#8ec5ff",
				Entrypoints: []string{"overview"}, ResourceHandlers: []string{"instance"},
				Capabilities: []string{"window.open"}, Dependencies: map[string]string{},
				WindowPolicy: "multiple", TabPolicy: &extensions.TabPolicy{Types: []string{"overview", "resource"}, Movable: true}, StateSchemaVersion: 1,
			},
			SHA256: hex.EncodeToString(digest[:]), Payload: payload,
		}
		if err := catalog.Publish(pkg); err != nil {
			return err
		}
	}
	return nil
}
