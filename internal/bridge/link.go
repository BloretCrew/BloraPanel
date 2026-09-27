// Package bridge carries bounded management RPCs on independent node channels.
package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

const StreamID = "rpc-v1"
const MaxFrameBytes = 128 * 1024
const MaxCalls = 8

type Request struct {
	ID       string                `json:"id"`
	Method   string                `json:"method"`
	ActorID  string                `json:"actorId"`
	Resource model.ResourceRef     `json:"resource"`
	Config   *model.InstanceConfig `json:"config,omitempty"`
	Args     json.RawMessage       `json:"args"`
}
type response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *model.APIError `json:"error,omitempty"`
}
type frame struct {
	Request  *Request  `json:"request,omitempty"`
	Response *response `json:"response,omitempty"`
}
type Handler func(context.Context, Request) (any, error)
type Link struct {
	conn    *protocol.Conn
	ctx     context.Context
	cancel  context.CancelFunc
	handler Handler
	mu      sync.Mutex
	pending map[string]chan response
	sendMu  sync.Mutex
	sent    uint64
	calls   chan struct{}
	workers chan Request
	wg      sync.WaitGroup
}

func New(ctx context.Context, c *protocol.Conn, handler Handler) *Link {
	ctx, cancel := context.WithCancel(ctx)
	l := &Link{conn: c, ctx: ctx, cancel: cancel, handler: handler, pending: map[string]chan response{}, calls: make(chan struct{}, MaxCalls), workers: make(chan Request, MaxCalls)}
	for range 4 {
		l.wg.Add(1)
		go func() { defer l.wg.Done(); l.worker() }()
	}
	l.wg.Add(1)
	go func() { defer l.wg.Done(); l.read() }()
	return l
}
func (l *Link) Done() <-chan struct{} { return l.ctx.Done() }
func (l *Link) Close() error          { l.cancel(); return l.conn.Close() }
func (l *Link) Wait()                 { l.wg.Wait() }

// Call sends once. Errors can leave an unknown remote outcome; mutation callers
// reconcile task/transfer IDs rather than automatically repeating side effects.
func (l *Link) Call(ctx context.Context, r Request, out any) error {
	select {
	case l.calls <- struct{}{}:
		defer func() { <-l.calls }()
	case <-ctx.Done():
		return ctx.Err()
	case <-l.ctx.Done():
		return protocol.ErrClosed
	}
	if r.ID == "" {
		r.ID = model.ID()
	}
	if r.Method == "" || r.ActorID == "" {
		return errors.New("RPC method and actor required")
	}
	result := make(chan response, 1)
	l.mu.Lock()
	if _, exists := l.pending[r.ID]; exists {
		l.mu.Unlock()
		return errors.New("duplicate active RPC identity")
	}
	l.pending[r.ID] = result
	l.mu.Unlock()
	defer func() { l.mu.Lock(); delete(l.pending, r.ID); l.mu.Unlock() }()
	if err := l.send(ctx, frame{Request: &r}); err != nil {
		return err
	}
	select {
	case reply := <-result:
		if reply.Error != nil {
			return reply.Error
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(reply.Result, out)
	case <-ctx.Done():
		return ctx.Err()
	case <-l.ctx.Done():
		return protocol.ErrClosed
	}
}
func (l *Link) send(ctx context.Context, f frame) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if len(b) > MaxFrameBytes {
		return errors.New("RPC frame exceeds 128 KiB budget")
	}
	l.sendMu.Lock()
	defer l.sendMu.Unlock()
	sequence := l.sent + uint64(len(b))
	for {
		err = l.conn.Send(ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: StreamID, Sequence: sequence, Payload: b})
		if !errors.Is(err, protocol.ErrNoCredit) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-l.ctx.Done():
			return protocol.ErrClosed
		case <-time.After(5 * time.Millisecond):
		}
	}
	if err == nil {
		l.sent = sequence
	}
	return err
}
func (l *Link) read() {
	defer l.Close()
	for {
		e, err := l.conn.Read(l.ctx)
		if err != nil {
			return
		}
		if e.Type != protocol.TypeData || e.StreamID != StreamID || len(e.Payload) > MaxFrameBytes {
			return
		}
		var f frame
		if err := json.Unmarshal(e.Payload, &f); err != nil {
			return
		}
		if (f.Request == nil) == (f.Response == nil) {
			return
		}
		if f.Request != nil {
			r := *f.Request
			if len(r.ID) > 128 || r.ID == "" || r.ActorID == "" || len(r.Method) > 100 {
				return
			}
			select {
			case l.workers <- r:
			case <-l.ctx.Done():
				return
			}
		} else {
			l.mu.Lock()
			ch := l.pending[f.Response.ID]
			l.mu.Unlock()
			if ch != nil {
				select {
				case ch <- *f.Response:
				default:
					return
				}
			}
		}
		// Receipt by this bounded RPC consumer is separate from downstream xterm
		// parse credit. A browser stream gates fetching its next batch on that credit.
		if err := l.conn.Consume(l.ctx, StreamID, e.Sequence); err != nil {
			return
		}
	}
}
func (l *Link) worker() {
	for {
		select {
		case <-l.ctx.Done():
			return
		case r := <-l.workers:
			ctx, cancel := context.WithTimeout(l.ctx, 20*time.Second)
			var value any
			var err error
			if l.handler == nil {
				err = &model.APIError{Code: "RPC_DIRECTION_DENIED", Message: "节点不能调用此方向的后台能力"}
			} else {
				value, err = l.handler(ctx, r)
			}
			reply := response{ID: r.ID}
			if err != nil {
				var api *model.APIError
				if errors.As(err, &api) {
					reply.Error = api
				} else {
					reply.Error = &model.APIError{Code: "NODE_OPERATION_FAILED", Message: err.Error()}
				}
			} else {
				reply.Result, err = json.Marshal(value)
				if err != nil {
					reply.Error = &model.APIError{Code: "ENCODING_ERROR", Message: err.Error()}
				}
			}
			if err := l.send(ctx, frame{Response: &reply}); err != nil {
				cancel()
				_ = l.Close()
				return
			}
			cancel()
		}
	}
}
