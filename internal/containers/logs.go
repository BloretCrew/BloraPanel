package containers

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"time"
)

// Logs reads Docker's retained timestamped output. It deliberately does not
// invent a terminal cursor: time/tail windows may overlap or have retention
// gaps. emit provides bounded backpressure; cancelling closes the Engine body.
func (m *Manager) Logs(ctx context.Context, actor string, target Target, options LogOptions, emit func(LogFrame) error) error {
	if err := m.authorize(ctx, actor); err != nil {
		return err
	}
	if err := m.requireEngine(); err != nil {
		return err
	}
	if emit == nil || options.Tail < 0 || options.Tail > 10000 {
		return ErrIdentity
	}
	o, err := m.engine.inspect(ctx, target, true)
	if err != nil {
		return err
	}
	if target.Kind != "container" {
		return ErrIdentity
	}
	var container containerInspect
	if err = json.Unmarshal(o.Details, &container); err != nil {
		return err
	}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if e := m.authorize(ctx, actor); e != nil {
					cancel(e)
					return
				}
			}
		}
	}()
	tail := options.Tail
	if tail == 0 {
		tail = 200
	}
	q := url.Values{"stdout": {"true"}, "stderr": {"true"}, "timestamps": {"true"}, "tail": {strconv.Itoa(tail)}, "follow": {strconv.FormatBool(options.Follow)}}
	if !options.Since.IsZero() {
		q.Set("since", options.Since.UTC().Format(time.RFC3339Nano))
	}
	res, err := m.engine.response(ctx, "GET", "/v"+apiVersion+"/containers/"+target.ID+"/logs?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	deliver := func(stream string, data []byte) error {
		if err := m.authorize(ctx, actor); err != nil {
			return err
		}
		return emit(LogFrame{Stream: stream, Data: append([]byte{}, data...)})
	}
	buffer := make([]byte, 32<<10)
	for {
		if err := ctx.Err(); err != nil {
			return context.Cause(ctx)
		}
		if container.Config.Tty {
			n, e := res.Body.Read(buffer)
			if n > 0 {
				if err = deliver("tty", buffer[:n]); err != nil {
					return err
				}
			}
			if errors.Is(e, io.EOF) {
				return nil
			}
			if e != nil {
				if ctx.Err() != nil {
					return context.Cause(ctx)
				}
				return e
			}
			continue
		}
		var header [8]byte
		if _, err = io.ReadFull(res.Body, header[:]); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			if ctx.Err() != nil {
				return context.Cause(ctx)
			}
			return err
		}
		if header[0] != 1 && header[0] != 2 {
			return ErrIdentity
		}
		remaining := int64(binary.BigEndian.Uint32(header[4:]))
		if remaining > 16<<20 {
			return errors.New("Docker log frame exceeds 16 MiB")
		}
		stream := "stdout"
		if header[0] == 2 {
			stream = "stderr"
		}
		for remaining > 0 {
			n := min(int64(len(buffer)), remaining)
			if _, err = io.ReadFull(res.Body, buffer[:n]); err != nil {
				return err
			}
			if err = deliver(stream, buffer[:n]); err != nil {
				return err
			}
			remaining -= n
		}
	}
}
