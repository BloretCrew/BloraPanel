package master

import (
	"encoding/json"

	"blora.dev/panel/internal/terminal"
)

const terminalFrameBytes = 64 << 10
const terminalFrameEvents = 128

type terminalDataFrame struct {
	payload []byte
	cursor  uint64
}

// Package only the archive events already returned by this read. There is no
// output coalescing timer and no change to archive cursors or byte credits.
// Legacy consumers retain their original single-event payloads.
func terminalDataFrames(events []terminal.Event, after uint64, grouped bool) ([]terminalDataFrame, error) {
	const prefix = `{"kind":"batch","events":[`
	var frames []terminalDataFrame
	var payload []byte
	var cursor uint64
	count := 0
	flush := func() {
		if count != 0 {
			payload = append(payload, ']', '}')
			frames = append(frames, terminalDataFrame{payload, cursor})
			payload = nil
			count = 0
		}
	}
	for _, event := range events {
		if event.Sequence <= after {
			continue
		}
		after = event.Sequence
		encoded, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}
		if !grouped {
			frames = append(frames, terminalDataFrame{encoded, event.Sequence})
			continue
		}
		if count != 0 && (count == terminalFrameEvents || len(payload)+1+len(encoded)+2 > terminalFrameBytes) {
			flush()
		}
		if len(prefix)+len(encoded)+2 > terminalFrameBytes {
			// Preserve an oversized legacy event intact, never truncate bytes.
			frames = append(frames, terminalDataFrame{encoded, event.Sequence})
			continue
		}
		if count == 0 {
			payload = append(payload, prefix...)
		} else {
			payload = append(payload, ',')
		}
		payload = append(payload, encoded...)
		cursor = event.Sequence
		count++
	}
	flush()
	return frames, nil
}
