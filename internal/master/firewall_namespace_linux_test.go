//go:build linux

package master

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/systeminfo"
	"golang.org/x/sys/unix"
)

// No firewall command is reachable before separate network and mount namespaces
// are verified and the host system bus is hidden by a private /run mount.
func TestPrivateFirewalldApplyAndRestore(t *testing.T) {
	if os.Getenv("BLORA_TEST_FIREWALL_NAMESPACE") != "1" {
		t.Skip("requires explicitly enabled private firewalld environment")
	}
	netNS, err := os.Readlink("/proc/self/ns/net")
	if err != nil {
		t.Fatal(err)
	}
	mountNS, err := os.Readlink("/proc/self/ns/mnt")
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("BLORA_FIREWALL_PARENT_NET") == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "unshare", "--user", "--map-root-user", "--mount", "--net", os.Args[0], "-test.run=^TestPrivateFirewalldApplyAndRestore$", "-test.v")
		cmd.Env = append(os.Environ(), "BLORA_FIREWALL_PARENT_NET="+netNS, "BLORA_FIREWALL_PARENT_MOUNT="+mountNS)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("private firewall environment: %v\n%s", err, output)
		}
		t.Logf("%s", output)
		return
	}
	if netNS == os.Getenv("BLORA_FIREWALL_PARENT_NET") || mountNS == os.Getenv("BLORA_FIREWALL_PARENT_MOUNT") {
		t.Fatal("refusing firewall test in parent namespace")
	}
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mount("blora-private-run", "/run", "tmpfs", unix.MS_NODEV|unix.MS_NOSUID, "size=8388608,mode=0755"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unix.Unmount("/run", unix.MNT_DETACH) })
	// The broker requires a journal socket. Keep diagnostic traffic inside this
	// namespace too; a private datagram sink is sufficient for its logging.
	if err := os.MkdirAll("/run/systemd/journal", 0700); err != nil {
		t.Fatal(err)
	}
	journal, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: "/run/systemd/journal/socket", Net: "unixgram"})
	if err != nil {
		t.Fatal(err)
	}
	journalDone := make(chan struct{})
	go func() {
		defer close(journalDone)
		buffer := make([]byte, 65536)
		for {
			if _, _, err := journal.ReadFromUnix(buffer); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() { _ = journal.Close(); <-journalDone })
	root := t.TempDir()
	socket := filepath.Join(root, "bus")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	fd, err := listener.File()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fd.Close() })
	config := filepath.Join(root, "bus.conf")
	if err := os.WriteFile(config, []byte(`<busconfig><type>system</type><policy context="default"><allow user="*"/><allow own="*"/><allow send_destination="*"/><allow receive_sender="*"/></policy></busconfig>`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DBUS_SYSTEM_BUS_ADDRESS", "unix:path="+socket)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path="+socket)
	start := func(name string, cmd *exec.Cmd) {
		t.Helper()
		logPath := filepath.Join(root, name+".log")
		log, err := os.Create(logPath)
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stdout, cmd.Stderr = log, log
		if err := cmd.Start(); err != nil {
			_ = log.Close()
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = cmd.Process.Signal(unix.SIGTERM)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				_ = cmd.Process.Kill()
				<-done
			}
			_ = log.Close()
			if t.Failed() {
				data, _ := os.ReadFile(logPath)
				t.Logf("%s: %s", name, data)
			}
		})
	}
	broker := exec.Command("sh", "-c", `export LISTEN_PID=$$; exec "$@"`, "sh", "dbus-broker-launch", "--scope", "user", "--config-file", config)
	broker.ExtraFiles = []*os.File{fd}
	broker.Env = append(os.Environ(), "LISTEN_FDS=1")
	start("broker", broker)
	privateConfig := filepath.Join(root, "firewalld")
	if err := os.Mkdir(privateConfig, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(privateConfig, "firewalld.conf"), []byte("DefaultZone=public\nFirewallBackend=nftables\nCleanupOnExit=yes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	start("firewalld", exec.Command("firewalld", "--nofork", "--nopid", "--system-config", privateConfig, "--log-target", "console"))
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var last []byte
	for {
		probe, stop := context.WithTimeout(ctx, time.Second)
		last, err = exec.CommandContext(probe, "firewall-cmd", "--state").CombinedOutput()
		stop()
		if err == nil && strings.TrimSpace(string(last)) == "running" {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("private firewalld not ready: %v %s", err, last)
		case <-time.After(100 * time.Millisecond):
		}
	}
	run := func(args ...string) string {
		t.Helper()
		out, err := exec.CommandContext(ctx, "firewall-cmd", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("private firewall-cmd %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("--add-port=1111/tcp")
	run("--add-port=1000-1002/tcp")
	run("--permanent", "--add-port=2222/udp")
	run("--permanent", "--add-port=2000-2002/udp")
	services := run("--list-services")
	before, err := systeminfo.CaptureFirewall(ctx)
	if err != nil {
		t.Fatal(err)
	}
	desired := []string{"3333/tcp"}
	if err := systeminfo.ApplyFirewallSnapshot(ctx, before, desired); err != nil {
		t.Fatal(err)
	}
	if run("--query-port=3333/tcp") != "yes" || run("--permanent", "--query-port=3333/tcp") != "yes" || run("--list-services") != services {
		t.Fatal("real apply lost rules or unrelated services")
	}
	rules, err := exec.CommandContext(ctx, "nft", "list", "ruleset").CombinedOutput()
	if err != nil || !strings.Contains(string(rules), "3333") {
		t.Fatalf("kernel rules do not confirm applied port: %v %s", err, rules)
	}
	if err := systeminfo.RestoreFirewallSnapshot(ctx, before, desired); err != nil {
		t.Fatal(err)
	}
	after, err := systeminfo.CaptureFirewall(ctx)
	if err != nil || !reflect.DeepEqual(before, after) || run("--list-services") != services {
		t.Fatalf("rollback did not preserve both configurations: %+v %+v %v", before, after, err)
	}
	// Run the production Master/Daemon lease against this real firewalld, then
	// drop only the private Master's port to prevent confirmation and reconnect.
	if out, err := exec.CommandContext(ctx, "ip", "link", "set", "lo", "up").CombinedOutput(); err != nil {
		t.Fatalf("private loopback: %v %s", err, out)
	}
	f := newFileFixture(t)
	node := f.instances[0].NodeID
	base := "/nodes/" + node + "/system/firewall"
	f.reader.request("POST", base+"/preview", map[string]any{"desired": desired}, model.ID(), 403)
	preview := f.admin.request("POST", base+"/preview", map[string]any{"desired": desired}, model.ID(), 200)
	var planHash string
	if err := json.Unmarshal(preview["planHash"], &planHash); err != nil || len(planHash) != 64 {
		t.Fatalf("missing bound firewall preview: %v", err)
	}
	task := parseFileTask(t, f.admin.request("POST", base+"/apply", map[string]any{"desired": desired, "planHash": planHash, "confirmUntil": time.Now().UTC().Add(12 * time.Second).Format(time.RFC3339)}, model.ID(), 202))
	eventually(t, 8*time.Second, func() bool {
		current, err := f.store.Task(ctx, task.ID)
		return err == nil && current.Phase == "awaiting_confirmation"
	})
	endpoint, err := url.Parse(f.admin.base)
	if err != nil || endpoint.Port() == "" {
		t.Fatal("missing private Master port")
	}
	nft := func(args ...string) {
		t.Helper()
		out, err := exec.CommandContext(ctx, "nft", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("private nft %v: %v %s", args, err, out)
		}
	}
	nft("add", "table", "inet", "blora_test_disconnect")
	t.Cleanup(func() { _ = exec.Command("nft", "delete", "table", "inet", "blora_test_disconnect").Run() })
	nft("add", "chain", "inet", "blora_test_disconnect", "output", "{ type filter hook output priority -300; policy accept; }")
	nft("add", "rule", "inet", "blora_test_disconnect", "output", "tcp", "dport", endpoint.Port(), "drop")
	if connection, err := net.DialTimeout("tcp", endpoint.Host, 500*time.Millisecond); err == nil {
		_ = connection.Close()
		t.Fatal("private management connection was not blocked")
	}
	eventually(t, 20*time.Second, func() bool {
		current, err := systeminfo.CaptureFirewall(ctx)
		return err == nil && reflect.DeepEqual(before, current)
	})
	nft("delete", "table", "inet", "blora_test_disconnect")
	task = f.awaitFile(t, task, model.Failed)
	if task.Phase != "confirmation_expired_rolled_back" {
		t.Fatalf("lost management connection did not report rollback: %+v", task)
	}
	// An explicitly confirmed lease must persist, including after the node
	// reopens its durable records. Receipt replay must not apply rules again.
	confirmed := parseFileTask(t, f.admin.request("POST", base+"/apply", map[string]any{"desired": desired, "planHash": planHash, "confirmUntil": time.Now().UTC().Add(20 * time.Second).Format(time.RFC3339)}, model.ID(), 202))
	eventually(t, 8*time.Second, func() bool {
		current, err := f.store.Task(ctx, confirmed.ID)
		return err == nil && current.Phase == "awaiting_confirmation"
	})
	confirmationKey := model.ID()
	body := map[string]string{"taskId": confirmed.ID}
	f.reader.request("POST", base+"/confirm", body, model.ID(), 403)
	f.admin.request("POST", base+"/confirm", body, confirmationKey, 200)
	confirmed = f.awaitFile(t, confirmed, model.Succeeded)
	if confirmed.Phase != "confirmed" {
		t.Fatalf("confirmed lease result missing: %+v", confirmed)
	}
	f.restartNode(0)
	f.admin.request("POST", base+"/confirm", body, confirmationKey, 200)
	f.admin.request("POST", base+"/confirm", body, model.ID(), 409)
	if run("--query-port=3333/tcp") != "yes" || run("--permanent", "--query-port=3333/tcp") != "yes" {
		t.Fatal("confirmed rules were reverted during node restart")
	}
	replayed, err := f.store.Task(ctx, confirmed.ID)
	if err != nil || replayed.Revision != confirmed.Revision {
		t.Fatalf("confirmation replay changed terminal task: %+v %v", replayed, err)
	}
	if err := systeminfo.RestoreFirewallSnapshot(ctx, before, desired); err != nil {
		t.Fatal(err)
	}
}
