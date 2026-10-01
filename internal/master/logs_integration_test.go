package master

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
)

func TestLogSlowConsumerDeadlineAndCursorReattach(t *testing.T) {
	f := newFileFixture(t)
	created := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[0].NodeID, "name": "slow-deadline", "config": model.InstanceConfig{Mode: "native", Command: nativeTestCommand(t, "log-markers"), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 1, Escalate: true}}, model.ID(), 201)
	var instance model.Instance
	if err := json.Unmarshal(created["instance"], &instance); err != nil {
		t.Fatal(err)
	}
	control := func(action string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+instance.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
	}
	f.awaitFile(t, control("start"), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, control("stop"), model.Succeeded) })
	current, err := f.store.Instance(context.Background(), instance.ID)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dial := func(after uint64) *protocol.Conn {
		ws, _, err := websocket.Dial(ctx, strings.Replace(f.admin.base, "https://", "wss://", 1)+"/api/v1/instances/"+instance.ID+"/logs/"+current.RunID+"/stream?sequence="+strconv.FormatUint(after, 10), &websocket.DialOptions{HTTPClient: f.admin.client, HTTPHeader: http.Header{"Origin": []string{f.admin.base}}})
		if err != nil {
			t.Fatal(err)
		}
		conn, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{current.RunID}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		return conn
	}
	conn := dial(0)
	var cursor uint64
	started := time.Now()
	for {
		frame, err := conn.Read(ctx)
		if err != nil {
			t.Fatal("slow-consumer diagnostic not received", err)
		}
		if frame.Type == protocol.TypeData {
			var event terminal.Event
			if err := json.Unmarshal(frame.Payload, &event); err != nil {
				t.Fatal(err)
			}
			cursor = event.Sequence // Deliberately do not return byte credit.
		}
		if frame.Type == protocol.TypeError {
			var diagnostic model.APIError
			if err := json.Unmarshal(frame.Payload, &diagnostic); err != nil {
				t.Fatal(err)
			}
			if diagnostic.Code != "SLOW_CONSUMER" {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			break
		}
	}
	if cursor == 0 || time.Since(started) < 44*time.Second || time.Since(started) > 55*time.Second {
		t.Fatalf("cursor=%d elapsed=%s", cursor, time.Since(started))
	}
	resumed := dial(cursor)
	for {
		frame, err := resumed.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if frame.Type == protocol.TypeError {
			t.Fatalf("resume error: %s", frame.Payload)
		}
		if frame.Type != protocol.TypeData {
			continue
		}
		var event terminal.Event
		if err := json.Unmarshal(frame.Payload, &event); err != nil {
			t.Fatal(err)
		}
		if event.Sequence <= cursor || !strings.Contains(string(event.Data), "second-marker") {
			t.Fatalf("incorrect cursor resume: %+v", event)
		}
		break
	}
	t.Logf("slow consumer error delivered after %s; original run cursor %d resumed", time.Since(started), cursor)
}

func TestConsoleInputHasOneDeliveryAndGracefulStopUsesIndependentStdin(t *testing.T) {
	f := newFileFixture(t)
	created := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[0].NodeID, "name": "console-input", "config": model.InstanceConfig{Mode: "native", Directory: f.roots[0], Command: nativeTestCommand(t, "console"), Environment: nativeTestEnvironment(), StopInput: "quit\n", StopSeconds: 1, KillSeconds: 1, Escalate: true}}, model.ID(), 201)
	var i model.Instance
	if err := json.Unmarshal(created["instance"], &i); err != nil {
		t.Fatal(err)
	}
	control := func(action string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
	}
	f.awaitFile(t, control("start"), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, control("stop"), model.Succeeded) })
	current, err := f.store.Instance(context.Background(), i.ID)
	if err != nil {
		t.Fatal(err)
	}
	url := "/instances/" + i.ID + "/logs/" + current.RunID + "/input"
	for _, action := range []string{"instance.read", "file.read"} {
		f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: action}, model.ID(), 200)
	}
	f.reader.request("POST", url, map[string]string{"data": "forbidden\n"}, model.ID(), 403)
	// Instance stdin is an independent grant: this user cannot create a host shell.
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "terminal.input"}, model.ID(), 200)
	key := model.ID()
	body := map[string]string{"data": "中文-one-delivery\n"}
	f.reader.request("POST", url, body, "", 400)
	first := parseFileTask(t, f.reader.request("POST", url, body, key, 202))
	second := parseFileTask(t, f.reader.request("POST", url, body, key, 202))
	if first.ID != second.ID {
		t.Fatal("same request created another console input task")
	}
	result := f.awaitFile(t, first, model.Succeeded)
	var delivered struct {
		RunID        string `json:"runId"`
		Bytes        int    `json:"bytesWritten"`
		DeliveryOnly bool   `json:"deliveryOnly"`
	}
	if err := json.Unmarshal(result.Result, &delivered); err != nil || delivered.RunID != current.RunID || delivered.Bytes != len(body["data"]) || !delivered.DeliveryOnly {
		t.Fatalf("stdin delivery receipt: %+v %v", delivered, err)
	}
	eventually(t, 3*time.Second, func() bool {
		b, err := os.ReadFile(filepath.Join(f.roots[0], "input-count"))
		return err == nil && string(b) == "x"
	})
	f.reader.request("POST", url, map[string]string{"data": "different\n"}, key, 409)
	f.awaitFile(t, control("stop"), model.Succeeded)
	assertDisk(t, f.roots[0], "graceful-stop", "stopped")
	assertDisk(t, f.roots[0], "input-count", "x")
	f.awaitFile(t, control("start"), model.Succeeded)
	f.reader.request("POST", url, map[string]string{"data": "stale-run\n"}, model.ID(), 409)
	f.awaitFile(t, control("stop"), model.Succeeded)
	assertDisk(t, f.roots[0], "input-count", "x")
}

func TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop(t *testing.T) {
	soakSeconds := 0
	if raw := os.Getenv("BLORA_A09_SOAK_SECONDS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 || parsed > 30 {
			t.Fatalf("BLORA_A09_SOAK_SECONDS must be an integer from 0 to 30, below the 45-second slow-consumer disconnect deadline; got %q", raw)
		}
		soakSeconds = parsed
	}
	producerDelay := "1ms"
	if soakSeconds > 0 {
		// Keep sustained pressure material but avoid generating gigabytes during
		// an opt-in long consumer stall.
		producerDelay = "10ms"
	}
	f := newFileFixture(t)
	created := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[0].NodeID, "name": "continuous-log", "config": model.InstanceConfig{Mode: "native", Command: nativeTestCommand(t, "continuous-log", producerDelay), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 1, Escalate: true}}, model.ID(), 201)
	var i model.Instance
	if err := json.Unmarshal(created["instance"], &i); err != nil {
		t.Fatal(err)
	}
	control := func(action string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
	}
	f.awaitFile(t, control("start"), model.Succeeded)
	t.Cleanup(func() {
		// Daemon shutdown preserves runs; stop our producer even on assertion failure.
		current, err := f.store.Instance(context.Background(), i.ID)
		if err == nil && current.State != "STOPPED" {
			f.awaitFile(t, control("stop"), model.Succeeded)
		}
	})
	current, err := f.store.Instance(context.Background(), i.ID)
	if err != nil {
		t.Fatal(err)
	}
	listed := f.admin.request("GET", "/instances/"+i.ID+"/logs", nil, "", 200)
	var logs []model.RunLog
	if err := json.Unmarshal(listed["items"], &logs); err != nil || len(logs) != 1 || logs[0].RunID != current.RunID {
		t.Fatalf("run log index: %+v %v", logs, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30+soakSeconds)*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, strings.Replace(f.admin.base, "https://", "wss://", 1)+"/api/v1/instances/"+i.ID+"/logs/"+current.RunID+"/stream?sequence=0", &websocket.DialOptions{HTTPClient: f.admin.client, HTTPHeader: http.Header{"Origin": []string{f.admin.base}}})
	if err != nil {
		t.Fatal(err)
	}
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{current.RunID}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	var first protocol.Envelope
	for {
		first, err = c.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if first.Type == protocol.TypeError {
			t.Fatalf("log stream: %s", first.Payload)
		}
		if first.Type == protocol.TypeData {
			break
		}
	}
	var event terminal.Event
	if err := json.Unmarshal(first.Payload, &event); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(event.Data), "中文日志-ready") {
		t.Fatalf("real process log missing: %q", event.Data)
	}
	// Hold the first byte credit: the browser is intentionally not consuming.
	eventually(t, 5*time.Second, func() bool {
		return c.Stats().Flow.ReceivedUnconsumed > 0
	})
	var peakReceived uint64
	peakQueue := 0
	for sample := 0; sample < 20; sample++ {
		stats := c.Stats()
		if stats.Flow.ReceivedUnconsumed > protocol.DefaultStreamWindow || stats.IncomingBytes > 2*protocol.MaxMessageSize || stats.IncomingQueued > 64 {
			t.Fatalf("slow reader exceeded connection budgets: %+v", stats)
		}
		peakReceived = max(peakReceived, stats.Flow.ReceivedUnconsumed)
		peakQueue = max(peakQueue, stats.IncomingBytes)
		time.Sleep(50 * time.Millisecond)
	}
	if soakSeconds > 0 {
		for sample := 0; sample < soakSeconds; sample++ {
			time.Sleep(time.Second)
			stats := c.Stats()
			if stats.Flow.ReceivedUnconsumed > protocol.DefaultStreamWindow || stats.IncomingBytes > 2*protocol.MaxMessageSize || stats.IncomingQueued > 64 {
				t.Fatalf("slow reader exceeded connection budgets during %ds soak: %+v", soakSeconds, stats)
			}
			peakReceived = max(peakReceived, stats.Flow.ReceivedUnconsumed)
			peakQueue = max(peakQueue, stats.IncomingBytes)
		}
		t.Logf("sustained stalled reader: duration=%ds samples=%d received-unconsumed peak=%d/%d incoming queue peak=%d/%d", soakSeconds, soakSeconds, peakReceived, protocol.DefaultStreamWindow, peakQueue, 2*protocol.MaxMessageSize)
	}
	t.Logf("stalled log reader: received-unconsumed peak=%d/%d, incoming queue peak=%d/%d", peakReceived, protocol.DefaultStreamWindow, peakQueue, 2*protocol.MaxMessageSize)
	// Keep real bulk traffic on the same source node while its log reader stalls.
	body := bytes.Repeat([]byte("bulk-with-stalled-log\n"), 400000)
	transferWrite(t, f.roots[0], "stalled-log-transfer.bin", body)
	transfer := startTransfer(t, f, "stalled-log-transfer.bin", "stalled-log-result.bin", false, model.ID())
	eventually(t, 15*time.Second, func() bool {
		task, progress := transferSnapshot(t, f, transfer.ID)
		return task.State == model.Running && progress.CurrentOffset > 0
	})
	started := time.Now()
	f.awaitFile(t, control("stop"), model.Succeeded)
	if time.Since(started) > 5*time.Second {
		t.Fatal("slow log consumer delayed control beyond stop/drain budget")
	}
	if err := c.Consume(ctx, current.RunID, first.Sequence); err != nil {
		t.Fatal(err)
	}
	current, err = f.store.Instance(context.Background(), i.ID)
	if err != nil || current.State != "STOPPED" {
		t.Fatal("actual run exit was not confirmed")
	}
	// A second reader can replay the persisted run after process exit. Reading
	// historical logs does not recreate either the instance or the log helper.
	var batch terminal.Batch
	args, _ := json.Marshal(map[string]any{"runId": current.RunID, "after": uint64(0), "maxBytes": 64 << 10})
	if err := f.app.nodeCall(ctx, i.NodeID, protocol.ChannelInteractive, bridge.Request{Method: "log.read", ActorID: f.adminUser.ID, Resource: instanceRef(i), Args: args}, &batch); err != nil {
		t.Fatal(err)
	}
	if len(batch.Events) == 0 || batch.Latest < event.Sequence {
		t.Fatal("finished run archive was lost")
	}
	result := awaitTransfer(t, f, transfer, model.Succeeded)
	if !result.DestinationVerified {
		t.Fatal("bulk transfer target was not verified")
	}
	written, err := os.ReadFile(filepath.Join(f.roots[1], "stalled-log-result.bin"))
	if err != nil || !bytes.Equal(written, body) {
		t.Fatal("concurrent bulk transfer content mismatch", err)
	}
}
