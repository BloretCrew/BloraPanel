package master

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"blora.dev/panel/internal/terminal"
)

func TestTerminalDataFramesPreserveOrderBytesAndBounds(t *testing.T) {
	var events []terminal.Event
	for index := 1; index <= 200; index++ {
		events = append(events, terminal.Event{Sequence: uint64(index), Kind: "output", Data: append([]byte{0, 255, 0xe4, '\x1b'}, bytes.Repeat([]byte("x"), 1024)...)})
	}
	events[7] = terminal.Event{Sequence: 8, Kind: "resize", Cols: 100, Rows: 28}
	for _, grouped := range []bool{false, true} {
		frames, err := terminalDataFrames(events, 3, grouped)
		if err != nil {
			t.Fatal(err)
		}
		var restored []terminal.Event
		var biggest int
		for _, frame := range frames {
			if len(frame.payload) > terminalFrameBytes {
				t.Fatal("payload budget exceeded")
			}
			var decoded []terminal.Event
			if grouped {
				var batch struct {
					Kind   string           `json:"kind"`
					Events []terminal.Event `json:"events"`
				}
				if err := json.Unmarshal(frame.payload, &batch); err != nil || batch.Kind != "batch" || len(batch.Events) == 0 || len(batch.Events) > terminalFrameEvents {
					t.Fatalf("invalid bounded batch: %v", err)
				}
				decoded = batch.Events
			} else {
				var event terminal.Event
				if err := json.Unmarshal(frame.payload, &event); err != nil {
					t.Fatal(err)
				}
				decoded = []terminal.Event{event}
			}
			if frame.cursor != decoded[len(decoded)-1].Sequence {
				t.Fatal("cursor does not identify last complete event")
			}
			biggest = max(biggest, len(decoded))
			restored = append(restored, decoded...)
		}
		if !reflect.DeepEqual(restored, events[3:]) {
			t.Fatal("event bytes, resize or sequence changed")
		}
		if grouped && (biggest <= 1 || len(frames) >= len(events)/4) {
			t.Fatal("ready events were not grouped")
		}
	}
}

func TestTerminalDataFramesBoundTinyEventsAndRetainLargeLegacyEvent(t *testing.T) {
	var events []terminal.Event
	for index := 1; index <= 260; index++ {
		events = append(events, terminal.Event{Sequence: uint64(index), Kind: "resize", Cols: 80, Rows: 24})
	}
	frames, err := terminalDataFrames(events, 0, true)
	if err != nil || len(frames) != 3 {
		t.Fatalf("event count bound: %d %v", len(frames), err)
	}
	large := terminal.Event{Sequence: 261, Kind: "output", Data: bytes.Repeat([]byte("x"), terminalFrameBytes)}
	frames, err = terminalDataFrames(append(events, large), 0, true)
	if err != nil {
		t.Fatal(err)
	}
	var decoded terminal.Event
	if err = json.Unmarshal(frames[len(frames)-1].payload, &decoded); err != nil || !reflect.DeepEqual(decoded, large) {
		t.Fatal("oversized legacy event was truncated or reordered")
	}
}
