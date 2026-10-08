package master

import (
	"context"
	"encoding/json"
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

func TestTerminalRealPTYGroupedOutputAndCursorRecovery(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native Linux PTY integration")
	}
	f := newFileFixture(t)
	instance := f.instances[0]
	eventually(t, 5*time.Second, func() bool {
		f.app.mu.Lock()
		defer f.app.mu.Unlock()
		return f.app.links[dataKey(instance.NodeID, protocol.ChannelInteractive)] != nil
	})
	created := f.awaitFile(t, parseFileTask(t, f.admin.request("POST", "/instances/"+instance.ID+"/terminals", map[string]int{"cols": 90, "rows": 24}, model.ID(), 202)), model.Succeeded)
	var result struct {
		Session terminal.Session `json:"session"`
	}
	if err := json.Unmarshal(created.Result, &result); err != nil {
		t.Fatal(err)
	}
	session := result.Session
	client := openTerminalEncoding(t, f.admin, session.ID, "grouped-native", 0, "events-v1")
	biggest := 0
	holdFirstAck := false
	read := func(c *terminalClient) protocol.Type {
		e, err := c.conn.Read(c.ctx)
		if err != nil {
			t.Fatal(err)
		}
		switch e.Type {
		case protocol.TypeError:
			t.Fatal("native terminal stream returned an error")
		case protocol.TypeOpenAck:
			var info struct {
				Writable bool `json:"writable"`
			}
			if err = json.Unmarshal(e.Payload, &info); err != nil {
				t.Fatal(err)
			}
			c.writable = info.Writable
		case protocol.TypeData:
			var data struct {
				Kind   string           `json:"kind"`
				Events []terminal.Event `json:"events"`
			}
			if err = json.Unmarshal(e.Payload, &data); err != nil || data.Kind != "batch" || len(data.Events) == 0 || len(data.Events) > terminalFrameEvents || len(e.Payload) > terminalFrameBytes {
				t.Fatal("invalid native grouped frame")
			}
			biggest = max(biggest, len(data.Events))
			for _, event := range data.Events {
				if event.Sequence != c.cursor+1 {
					t.Fatal("grouped archive cursor gap")
				}
				c.cursor = event.Sequence
				if event.Kind == "output" {
					c.output.Write(event.Data)
				}
			}
			if holdFirstAck {
				// Accumulate genuinely committed PTY output behind destination
				// backpressure; production must not add a batching delay.
				holdFirstAck = false
				time.Sleep(140 * time.Millisecond)
			}
			if err = c.conn.Consume(c.ctx, c.id, e.Sequence); err != nil {
				t.Fatal(err)
			}
		}
		return e.Type
	}
	for read(client) != protocol.TypeResume {
	}
	if !client.writable {
		t.Fatal("native input lease missing")
	}
	holdFirstAck = true
	client.input("printf x >> batch-input-count; i=0; while [ $i -lt 12 ]; do printf '%01024d\\n' $i; i=$((i+1)); sleep 0.01; done; printf '\\033[32m\\344\\270\\255\\346\\226\\207-\\102\\101\\124\\103\\110-\\117\\113\\033[0m\\n'\n")
	for !strings.Contains(client.output.String(), "中文-BATCH-OK") {
		read(client)
	}
	if biggest <= 1 || !strings.Contains(client.output.String(), "\x1b[32m") {
		t.Fatal("ready output was not grouped or ANSI bytes changed")
	}
	cursor := client.cursor
	_ = client.conn.Close()
	refreshed := openTerminalEncoding(t, f.admin, session.ID, "grouped-native", cursor, "events-v1")
	for read(refreshed) != protocol.TypeResume {
	}
	data, err := os.ReadFile(filepath.Join(f.roots[0], "batch-input-count"))
	if err != nil || string(data) != "x" {
		t.Fatal("refresh repeated terminal input")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ws, response, err := websocket.Dial(ctx, strings.Replace(f.admin.base, "https://", "wss://", 1)+"/api/v1/terminals/"+session.ID+"/stream?viewId=invalid-encoding&sequence=0&encoding=unknown", &websocket.DialOptions{HTTPClient: f.admin.client, HTTPHeader: http.Header{"Origin": []string{f.admin.base}}})
	if ws != nil {
		_ = ws.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != 400 {
		t.Fatal("unsupported encoding was accepted")
	}
}
