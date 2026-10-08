package master

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
)

type terminalClient struct {
	t            *testing.T
	ctx          context.Context
	conn         *protocol.Conn
	id           string
	sent, cursor uint64
	output       strings.Builder
	writable     bool
}

func openTerminal(t *testing.T, c *testClient, id, view string, cursor uint64) *terminalClient {
	return openTerminalEncoding(t, c, id, view, cursor, "")
}
func openTerminalEncoding(t *testing.T, c *testClient, id, view string, cursor uint64, encoding string) *terminalClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	header := http.Header{"Origin": []string{c.base}}
	endpoint := fmt.Sprintf("/api/v1/terminals/%s/stream?viewId=%s&sequence=%d", id, view, cursor)
	if encoding != "" {
		endpoint += "&encoding=" + encoding
	}
	ws, _, err := websocket.Dial(ctx, strings.Replace(c.base, "https://", "wss://", 1)+endpoint, &websocket.DialOptions{HTTPClient: c.client, HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{id}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &terminalClient{t: t, ctx: ctx, conn: conn, id: id, cursor: cursor}
}
func (c *terminalClient) read() protocol.Type {
	c.t.Helper()
	e, err := c.conn.Read(c.ctx)
	if err != nil {
		c.t.Fatal(err)
	}
	switch e.Type {
	case protocol.TypeError:
		c.t.Fatalf("terminal stream error: %s", e.Payload)
	case protocol.TypeOpenAck:
		var ready struct {
			Writable bool `json:"writable"`
		}
		if err := json.Unmarshal(e.Payload, &ready); err != nil {
			c.t.Fatal(err)
		}
		c.writable = ready.Writable
	case protocol.TypeData:
		var event terminal.Event
		if err := json.Unmarshal(e.Payload, &event); err != nil {
			c.t.Fatal(err)
		}
		if event.Sequence != c.cursor+1 {
			c.t.Fatalf("event gap/duplicate: %d after %d", event.Sequence, c.cursor)
		}
		c.cursor = event.Sequence
		if event.Kind == "output" {
			c.output.Write(event.Data)
		}
		if err := c.conn.Consume(c.ctx, c.id, e.Sequence); err != nil {
			c.t.Fatal(err)
		}
	}
	return e.Type
}
func (c *terminalClient) resume() {
	c.t.Helper()
	for c.read() != protocol.TypeResume {
	}
}
func (c *terminalClient) input(value string) {
	c.t.Helper()
	c.sent += uint64(len(value))
	if err := c.conn.Send(c.ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: c.id, Sequence: c.sent, Payload: []byte(value)}); err != nil {
		c.t.Fatal(err)
	}
}

func TestTerminalRealPTYRefreshLeaseAndRevocation(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("real Linux PTY integration; Windows has separate executable acceptance")
	}
	f := newFileFixture(t)
	i := f.instances[0]
	eventually(t, 5*time.Second, func() bool {
		f.app.mu.Lock()
		defer f.app.mu.Unlock()
		return f.app.links[dataKey(i.NodeID, protocol.ChannelInteractive)] != nil
	})
	endpoint := "/instances/" + i.ID + "/terminals"
	// A file/read-only member cannot obtain a native host shell.
	f.reader.request("POST", endpoint, map[string]int{"cols": 90, "rows": 24}, model.ID(), 403)
	key := model.ID()
	created := parseFileTask(t, f.admin.request("POST", endpoint, map[string]int{"cols": 90, "rows": 24}, key, 202))
	duplicate := parseFileTask(t, f.admin.request("POST", endpoint, map[string]int{"cols": 90, "rows": 24}, key, 202))
	if duplicate.ID != created.ID {
		t.Fatal("terminal creation retry produced another task")
	}
	created = f.awaitFile(t, created, model.Succeeded)
	var result struct {
		Session terminal.Session `json:"session"`
	}
	if err := json.Unmarshal(created.Result, &result); err != nil {
		t.Fatal(err)
	}
	session := result.Session
	if session.ID == "" || session.Resource != instanceRef(i) || session.State != "running" {
		t.Fatalf("invalid actual session: %+v", session)
	}
	first := openTerminal(t, f.admin, session.ID, "pty-primary", 0)
	first.resume()
	if !first.writable {
		t.Fatal("creator did not receive input lease")
	}
	// Escaped printf proves shell execution instead of merely finding echo of
	// the command; a file records how many times the side effect really ran.
	first.input("printf x >> pty-input-count; printf '\\033[32m\\344\\270\\255\\346\\226\\207-PTY-OK\\033[0m\\n'\n")
	for !strings.Contains(first.output.String(), "中文-PTY-OK") {
		first.read()
	}
	if !strings.Contains(first.output.String(), "\x1b[32m") {
		t.Fatal("ANSI output was not preserved")
	}
	checkpoint := first.cursor
	_ = first.conn.Close()
	restored := openTerminal(t, f.admin, session.ID, "pty-primary", checkpoint)
	restored.resume()
	data, err := os.ReadFile(filepath.Join(f.roots[0], "pty-input-count"))
	if err != nil || string(data) != "x" {
		t.Fatalf("refresh repeated input: %q %v", data, err)
	}
	listed := f.admin.request("GET", endpoint, nil, "", 200)
	var sessions []terminal.Session
	if err := json.Unmarshal(listed["items"], &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatal("reattach replaced the original PTY")
	}
	for _, grant := range []model.Grant{
		{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "terminal.read"},
		{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "node", ID: i.NodeID}, Action: "host.manage"},
	} {
		f.admin.request("POST", "/grants", grant, model.ID(), 200)
	}
	observer := openTerminal(t, f.reader, session.ID, "pty-observer", 0)
	observer.resume()
	if observer.writable {
		t.Fatal("read-only user acquired input lease")
	}
	observer.input("printf forbidden >> pty-input-count\n")
	e, err := observer.conn.Read(observer.ctx)
	if err == nil && e.Type != protocol.TypeError {
		t.Fatalf("read-only input was accepted: %v", e.Type)
	}
	data, err = os.ReadFile(filepath.Join(f.roots[0], "pty-input-count"))
	if err != nil || string(data) != "x" {
		t.Fatal("read-only input caused a side effect")
	}
	watcher := openTerminal(t, f.reader, session.ID, "pty-watch-revocation", 0)
	watcher.resume()
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "terminal.read"}, model.ID(), 200)
	eventually(t, 3*time.Second, func() bool {
		select {
		case <-watcher.conn.Done():
			return true
		default:
			return false
		}
	})
	f.reader.request("GET", endpoint, nil, "", 403)
	closed := parseFileTask(t, f.admin.request("POST", "/terminals/"+session.ID+"/close", map[string]any{}, model.ID(), 202))
	f.awaitFile(t, closed, model.Succeeded)
	listed = f.admin.request("GET", endpoint, nil, "", 200)
	if err := json.Unmarshal(listed["items"], &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].State == "running" {
		t.Fatalf("session close was not observed: %+v", sessions)
	}
}
