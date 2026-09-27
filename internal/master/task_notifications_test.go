package master

import (
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestTaskNotificationStreamCursorAuthorizationAndResume(t *testing.T) {
	f := newFileFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	finish := func(actor string) model.Task {
		task, _, err := f.store.Accept(ctx, model.Task{ActorID: actor, RequestID: model.ID(), Resource: instanceRef(f.instances[0]), Action: "fixture.notification", Payload: json.RawMessage(`{"private":"must not appear"}`)})
		if err != nil {
			t.Fatal(err)
		}
		task, err = f.store.UpdateTask(ctx, task.ID, task.Revision, model.Failed, "fixture", nil, "private failure")
		if err != nil {
			t.Fatal(err)
		}
		return task
	}
	finish(f.readerUser.ID)
	connect := func(after string) *websocket.Conn {
		header := http.Header{"Origin": []string{f.reader.base}}
		req, _ := http.NewRequest("GET", f.reader.base, nil)
		for _, cookie := range f.reader.client.Jar.Cookies(req.URL) {
			header.Add("Cookie", cookie.String())
		}
		url := "wss" + f.reader.base[5:] + "/api/v1/events?notifications=1"
		if after != "" {
			url += "&after=" + after
		}
		ws, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPClient: f.reader.client, HTTPHeader: header})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { ws.CloseNow() })
		return ws
	}
	read := func(ws *websocket.Conn) protocol.Envelope {
		_, b, err := ws.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		frame, err := protocol.Unmarshal(b)
		if err != nil {
			t.Fatal(err)
		}
		return frame
	}
	ws := connect("")
	initial := read(ws)
	if initial.Type != protocol.TypeSnapshot || initial.Sequence == 0 {
		t.Fatal("missing initial boundary")
	}
	hidden := finish(f.adminUser.ID)
	wanted := finish(f.readerUser.ID)
	var terminalSequence uint64
	for terminalSequence == 0 {
		frame := read(ws)
		if frame.Type != protocol.TypeTaskEvent {
			continue
		}
		var value struct {
			Task map[string]json.RawMessage `json:"task"`
		}
		if err := json.Unmarshal(frame.Payload, &value); err != nil {
			t.Fatal(err)
		}
		if string(value.Task["taskId"]) != strconv.Quote(wanted.ID) || value.Task["payload"] != nil || value.Task["error"] != nil {
			t.Fatalf("unexpected notification, hidden task %s", hidden.ID)
		}
		terminalSequence = frame.Sequence
	}
	ws.CloseNow()
	resumed := connect(strconv.FormatUint(terminalSequence, 10))
	if frame := read(resumed); frame.Type != protocol.TypeSnapshot || frame.Sequence != terminalSequence {
		t.Fatal("resume boundary mismatch")
	}
	newest := finish(f.readerUser.ID)
	for {
		frame := read(resumed)
		if frame.Type != protocol.TypeTaskEvent {
			continue
		}
		var value struct {
			Task model.Task `json:"task"`
		}
		json.Unmarshal(frame.Payload, &value)
		if value.Task.ID != newest.ID || frame.Sequence <= terminalSequence {
			t.Fatal("replayed old terminal notification")
		}
		break
	}
}
