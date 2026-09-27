package master

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

func TestNodeKeyRotationKeepsIdentityAndFencesPreviousConnections(t *testing.T) {
	f := newFileFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	i := f.instances[0]
	_, oldKey, err := f.store.Node(ctx, i.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	f.app.mu.Lock()
	old := f.app.peers[i.NodeID]
	oldData := f.app.links[dataKey(i.NodeID, protocol.ChannelBulk)]
	f.app.mu.Unlock()
	f.reader.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{}, model.ID(), 403)
	f.admin.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{}, "", 400)
	rotationRequestID := model.ID()
	issued := f.admin.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{}, rotationRequestID, 201)
	replayed := f.admin.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{}, rotationRequestID, 201)
	if string(issued["ticket"]) != string(replayed["ticket"]) {
		t.Fatal("rotation ticket retry did not replay the original receipt")
	}
	f.admin.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{"reactivate": true}, rotationRequestID, 409)
	var ticket string
	if err := json.Unmarshal(issued["ticket"], &ticket); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "rotation.json")
	body, err := json.Marshal(map[string]string{"nodeId": i.NodeID, "ticket": ticket})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, body, 0600); err != nil {
		t.Fatal(err)
	}
	base := filepath.Dir(f.roots[0])
	config := daemon.Config{StateDir: filepath.Join(base, "files-node-a"), CAFile: filepath.Join(base, "ca.crt"), MasterURL: f.admin.base, AllowPGIDFallback: true, RotationFile: file}
	replacement, err := daemon.New(config)
	if err != nil {
		t.Fatal(err)
	}
	if replacement.NodeID() != i.NodeID {
		t.Fatal("rotation replaced durable node identity")
	}
	_, newKey, err := f.store.Node(ctx, i.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(oldKey, newKey) {
		t.Fatal("node key did not change")
	}
	eventually(t, 3*time.Second, func() bool {
		select {
		case <-old.conn.Done():
			return true
		default:
			return false
		}
	})
	eventually(t, 3*time.Second, func() bool {
		select {
		case <-oldData.Done():
			return true
		default:
			return false
		}
	})
	done := make(chan struct{})
	go func() { defer close(done); _ = replacement.Run(ctx) }()
	t.Cleanup(func() { cancel(); _ = replacement.Close(); <-done })
	eventually(t, 8*time.Second, func() bool {
		f.app.mu.Lock()
		defer f.app.mu.Unlock()
		p := f.app.peers[i.NodeID]
		return p != nil && p.generation > old.generation && f.app.links[dataKey(i.NodeID, protocol.ChannelBulk)] != nil
	})
	if _, err := f.store.Instance(ctx, i.ID); err != nil {
		t.Fatal("rotation lost associated instance")
	}
	// Replay of a confirmed ticket must not rotate again or close a new peer.
	if err := replacement.Close(); err != nil {
		t.Fatal(err)
	}
	<-done
	again, err := daemon.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	_, sameKey, err := f.store.Node(ctx, i.NodeID)
	if err != nil || !bytes.Equal(sameKey, newKey) {
		t.Fatal("retry generated another replacement key")
	}
	otherPublic, otherPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	f.admin.request("POST", "/agent/rotate", map[string]any{"nodeId": i.NodeID, "ticket": ticket, "publicKey": otherPublic, "signature": ed25519.Sign(otherPrivate, model.RotationProof(i.NodeID, ticket, otherPublic))}, "", 403)
	// A second explicit rotation can replace the previously committed pending
	// record; an ambiguous unfinished pending key would instead remain protected.
	second := f.admin.request("POST", "/nodes/"+i.NodeID+"/rotation", map[string]any{}, model.ID(), 201)
	if err := json.Unmarshal(second["ticket"], &ticket); err != nil {
		t.Fatal(err)
	}
	body, err = json.Marshal(map[string]string{"nodeId": i.NodeID, "ticket": ticket})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := again.Close(); err != nil {
		t.Fatal(err)
	}
	secondReplacement, err := daemon.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer secondReplacement.Close()
	_, thirdKey, err := f.store.Node(ctx, i.NodeID)
	if err != nil || bytes.Equal(thirdKey, newKey) {
		t.Fatal("subsequent rotation failed")
	}
}
