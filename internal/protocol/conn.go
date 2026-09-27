package protocol

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Options struct {
	Generation uint64
	Channel    Channel
	// CurrentGeneration must be concurrency safe. It fences replaced sessions.
	CurrentGeneration func() uint64
	QueueSize         int
	QueueBytes        int
	WriteTimeout      time.Duration
	// BulkBytesPerSecond limits this physical bulk connection. Zero selects
	// 8 MiB/s; its token bucket allows at most one message of initial burst.
	BulkBytesPerSecond uint64
	MaxStreams         int
	StreamWindow       uint64
	ConnectionWindow   uint64
	// Streams negotiated by the authenticated handshake, installed before reads.
	InitialStreams []string
}

type queuedFrame struct {
	ctx   context.Context
	bytes []byte
	done  chan error
}

// Conn owns exactly one already authenticated WSS connection. Handshake TLS,
// Origin and credential checks happen before NewConn in the HTTP service.
// Read is single-consumer. Send may be concurrent; queues and windows are bounded.
type Conn struct {
	ws            *websocket.Conn
	options       Options
	ctx           context.Context
	cancel        context.CancelCauseFunc
	closeOnce     sync.Once
	high, normal  chan queuedFrame
	incoming      chan Envelope
	enqueueMu     sync.Mutex
	budgetMu      sync.Mutex
	queuedBytes   int
	incomingBytes int
	flow          *FlowLedger
	rateTokens    float64
	rateLast      time.Time
}

func NewConn(ctx context.Context, ws *websocket.Conn, o Options) (*Conn, error) {
	if ws == nil {
		return nil, ErrClosed
	}
	if o.Generation == 0 {
		return nil, ErrGeneration
	}
	if !o.Channel.Valid() {
		return nil, ErrChannel
	}
	if o.QueueSize <= 0 {
		o.QueueSize = 64
	}
	if o.QueueSize > 1024 {
		return nil, ErrQueueFull
	}
	if o.QueueBytes <= 0 {
		o.QueueBytes = 2 * MaxMessageSize
	}
	if o.QueueBytes > 32*MaxMessageSize {
		return nil, ErrQueueFull
	}
	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 5 * time.Second
	}
	if o.BulkBytesPerSecond == 0 {
		o.BulkBytesPerSecond = 8 * 1024 * 1024
	}
	c := &Conn{ws: ws, options: o, high: make(chan queuedFrame, o.QueueSize), normal: make(chan queuedFrame, o.QueueSize), incoming: make(chan Envelope, o.QueueSize), flow: NewFlowLedger(o.MaxStreams, o.StreamWindow, o.ConnectionWindow)}
	c.ctx, c.cancel = context.WithCancelCause(ctx)
	for _, id := range o.InitialStreams {
		window := o.StreamWindow
		if window == 0 {
			window = DefaultStreamWindow
		}
		if err := c.flow.Open(id, window, window, 0, 0); err != nil {
			c.cancel(err)
			return nil, err
		}
	}
	c.rateTokens = MaxMessageSize
	c.rateLast = time.Now()
	limit := MaxMessageSize
	if o.Channel == ChannelControl {
		limit = MaxControlMessageSize
	}
	ws.SetReadLimit(int64(limit))
	go c.writer()
	go c.reader()
	go func() { <-c.ctx.Done(); c.closeSocket() }()
	return c, nil
}

func (c *Conn) Generation() uint64       { return c.options.Generation }
func (c *Conn) Channel() Channel         { return c.options.Channel }
func (c *Conn) Flow() *FlowLedger        { return c.flow }
func (c *Conn) Done() <-chan struct{}    { return c.ctx.Done() }
func (c *Conn) Err() error               { return context.Cause(c.ctx) }
func (c *Conn) Close() error             { c.cancel(ErrClosed); return c.closeSocket() }
func (c *Conn) closeSocket() (err error) { c.closeOnce.Do(func() { err = c.ws.CloseNow() }); return }
func (c *Conn) fail(err error)           { c.cancel(err) }
func (c *Conn) current() error {
	if err := context.Cause(c.ctx); err != nil {
		return err
	}
	if c.options.CurrentGeneration != nil && c.options.CurrentGeneration() != c.options.Generation {
		c.fail(ErrGeneration)
		return ErrGeneration
	}
	return nil
}

// Send returns only after a socket write. A non-nil error after submission can
// mean an unknown remote outcome; services reconcile requestId and never replay
// terminal input. Saturated queues/credit fail before submission with backpressure.
func (c *Conn) Send(ctx context.Context, e Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.current(); err != nil {
		return err
	}
	if e.ProtocolVersion == 0 {
		e.ProtocolVersion = Version
	}
	if e.Generation == 0 {
		e.Generation = c.options.Generation
	}
	if e.Channel == 0 {
		e.Channel = c.options.Channel
	}
	if e.Generation != c.options.Generation {
		return ErrGeneration
	}
	if e.Channel != c.options.Channel {
		return ErrChannel
	}
	b, err := Marshal(e)
	if err != nil {
		return err
	}
	q := queuedFrame{ctx: ctx, bytes: b, done: make(chan error, 1)}
	queue := c.normal
	if priorityMessage(e) {
		queue = c.high
	}
	c.enqueueMu.Lock()
	c.budgetMu.Lock()
	// Reserve half of each connection's byte budget for high-priority messages.
	limit := c.options.QueueBytes
	if e.Type == TypeData {
		limit /= 2
	}
	if c.queuedBytes+len(b) > limit {
		c.budgetMu.Unlock()
		c.enqueueMu.Unlock()
		return ErrQueueFull
	}
	if e.Type == TypeData {
		if err = c.flow.ReserveSend(e.StreamID, e.Sequence, uint64(len(e.Payload))); err != nil {
			c.budgetMu.Unlock()
			c.enqueueMu.Unlock()
			return err
		}
	}
	c.queuedBytes += len(b)
	select {
	case queue <- q:
	default:
		c.queuedBytes -= len(b)
		if e.Type == TypeData {
			c.flow.refundSend(e.StreamID, e.Sequence, uint64(len(e.Payload)))
		}
		c.budgetMu.Unlock()
		c.enqueueMu.Unlock()
		return ErrQueueFull
	}
	c.budgetMu.Unlock()
	c.enqueueMu.Unlock()
	select {
	case err := <-q.done:
		return err
	case <-ctx.Done():
		c.fail(ctx.Err())
		return ctx.Err()
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	}
}

func (c *Conn) Read(ctx context.Context) (Envelope, error) {
	if err := c.current(); err != nil {
		return Envelope{}, err
	}
	select {
	case e := <-c.incoming:
		c.budgetMu.Lock()
		c.incomingBytes -= envelopeMemory(e)
		c.budgetMu.Unlock()
		if err := c.current(); err != nil {
			return Envelope{}, err
		}
		return e, nil
	case <-ctx.Done():
		return Envelope{}, ctx.Err()
	case <-c.ctx.Done():
		return Envelope{}, context.Cause(c.ctx)
	}
}

func (c *Conn) Consume(ctx context.Context, streamID string, end uint64) error {
	credit, err := c.flow.Consumed(streamID, end)
	if err != nil {
		return err
	}
	err = c.Send(ctx, Envelope{Type: TypeAck, StreamID: streamID, Sequence: end, Credit: credit})
	if err != nil {
		c.fail(err)
	}
	return err
}

func (c *Conn) writer() {
	for {
		var q queuedFrame
		// Check high queue first, including while DATA producers remain active.
		select {
		case q = <-c.high:
		default:
			select {
			case q = <-c.high:
			case q = <-c.normal:
			case <-c.ctx.Done():
				return
			}
		}
		if err := c.current(); err != nil {
			q.done <- err
			return
		}
		ctx, cancel := context.WithTimeout(c.ctx, c.options.WriteTimeout)
		stop := context.AfterFunc(q.ctx, cancel)
		if q.ctx.Err() != nil {
			cancel()
		}
		err := c.awaitRate(ctx, len(q.bytes))
		if err == nil {
			err = c.ws.Write(ctx, websocket.MessageBinary, q.bytes)
		}
		stop()
		cancel()
		c.budgetMu.Lock()
		c.queuedBytes -= len(q.bytes)
		c.budgetMu.Unlock()
		q.done <- err
		if err != nil {
			c.fail(err)
			return
		}
	}
}

func priorityMessage(e Envelope) bool {
	if e.Channel == ChannelControl {
		return true
	}
	switch e.Type {
	case TypeAck, TypeCredit, TypePing, TypePong, TypeReset, TypeError:
		return true
	}
	return false
}

// Only the writer touches the bucket. A slow bulk destination cannot consume
// tokens or queue slots belonging to control/interactive physical connections.
func (c *Conn) awaitRate(ctx context.Context, size int) error {
	if c.options.Channel != ChannelBulk {
		return nil
	}
	for {
		now := time.Now()
		c.rateTokens += now.Sub(c.rateLast).Seconds() * float64(c.options.BulkBytesPerSecond)
		c.rateLast = now
		if c.rateTokens > MaxMessageSize {
			c.rateTokens = MaxMessageSize
		}
		if c.rateTokens >= float64(size) {
			c.rateTokens -= float64(size)
			return ctx.Err()
		}
		delay := time.Duration((float64(size) - c.rateTokens) / float64(c.options.BulkBytesPerSecond) * float64(time.Second))
		if delay < time.Millisecond {
			delay = time.Millisecond
		}
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (c *Conn) reader() {
	for {
		typ, b, err := c.ws.Read(c.ctx)
		if err != nil {
			c.fail(err)
			return
		}
		if err = c.current(); err != nil {
			return
		}
		if typ != websocket.MessageBinary {
			c.fail(ErrMalformed)
			return
		}
		e, err := Unmarshal(b)
		if err != nil {
			c.fail(err)
			return
		}
		if e.Generation != c.options.Generation {
			c.fail(ErrGeneration)
			return
		}
		if e.Channel != c.options.Channel {
			c.fail(ErrChannel)
			return
		}
		switch e.Type {
		case TypeData:
			err = c.flow.ReceiveData(e.StreamID, e.Sequence, uint64(len(e.Payload)))
		case TypeAck:
			err = c.flow.ReceiveAck(e.StreamID, e.Sequence, e.Credit)
		case TypeCredit:
			err = c.flow.ReceiveCredit(e.StreamID, e.Credit)
		}
		if err != nil {
			c.fail(err)
			return
		}
		// ACK/CREDIT are ledger events; don't let application read slowness hold them.
		if e.Type == TypeAck || e.Type == TypeCredit {
			continue
		}
		c.budgetMu.Lock()
		size := envelopeMemory(e)
		if c.incomingBytes+size > c.options.QueueBytes {
			c.budgetMu.Unlock()
			c.fail(ErrQueueFull)
			return
		}
		c.incomingBytes += size
		if c.options.Channel == ChannelControl {
			// Control messages have no DATA credit loop to service. Preserve
			// accepted commands during short storage stalls by applying socket
			// backpressure instead of disconnecting when the bounded queue fills.
			// Account for this one pending frame, and release the lock so Read
			// can drain the queue. Interactive/bulk readers remain nonblocking.
			c.budgetMu.Unlock()
			select {
			case c.incoming <- e:
			case <-c.ctx.Done():
				c.budgetMu.Lock()
				c.incomingBytes -= size
				c.budgetMu.Unlock()
				return
			}
			continue
		}
		select {
		case c.incoming <- e:
			c.budgetMu.Unlock()
		case <-c.ctx.Done():
			c.incomingBytes -= size
			c.budgetMu.Unlock()
			return
		default:
			c.incomingBytes -= size
			c.budgetMu.Unlock()
			c.fail(ErrQueueFull)
			return
		}
	}
}

func envelopeMemory(e Envelope) int {
	return len(e.Payload) + len(e.StreamID) + len(e.RequestID) + len(e.RunID) + 128
}

type ConnStats struct {
	QueuedBytes    int
	IncomingBytes  int
	HighQueued     int
	NormalQueued   int
	IncomingQueued int
	Flow           FlowStats
}

func (c *Conn) Stats() ConnStats {
	c.budgetMu.Lock()
	n, m := c.queuedBytes, c.incomingBytes
	c.budgetMu.Unlock()
	return ConnStats{n, m, len(c.high), len(c.normal), len(c.incoming), c.flow.Stats()}
}

// IsBackpressure means a message was not submitted; the producer can retain it
// within its own bounded buffer or use its archive/resume path.
func IsBackpressure(err error) bool {
	return errors.Is(err, ErrNoCredit) || errors.Is(err, ErrQueueFull)
}
