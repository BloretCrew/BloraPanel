package protocol

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestProtobufInteroperability(t *testing.T) {
	// Dynamic protobuf messages use the official full-message decoder, independently
	// of our constrained raw decoder. Verify descriptor fields against shipped .proto.
	names := []string{"protocol_version", "generation", "channel", "request_id", "run_id", "stream_id", "type", "sequence", "payload", "credit"}
	types := []descriptorpb.FieldDescriptorProto_Type{13, 4, 14, 9, 9, 9, 14, 4, 12, 4}
	schema, err := os.ReadFile("../../api/management.proto")
	if err != nil {
		t.Fatal(err)
	}
	md := &descriptorpb.DescriptorProto{Name: proto.String("Envelope")}
	typeNames := []string{"uint32", "uint64", "Channel", "string", "string", "string", "MessageType", "uint64", "bytes", "uint64"}
	for i, name := range names {
		if !strings.Contains(string(schema), fmt.Sprintf("%s %s = %d;", typeNames[i], name, i+1)) {
			t.Fatalf(".proto drift: %s", name)
		}
		f := &descriptorpb.FieldDescriptorProto{Name: proto.String(name), Number: proto.Int32(int32(i + 1)), Type: types[i].Enum(), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()}
		// An enum has the same protobuf wire type as uint32; use a uint32 descriptor
		// to independently check numeric enum values without recreating an enum list.
		if types[i] == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			f.Type = descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum()
		}
		md.Field = append(md.Field, f)
	}
	fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{Name: proto.String("interop.proto"), Syntax: proto.String("proto3"), MessageType: []*descriptorpb.DescriptorProto{md}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	original := Envelope{Version, 23, ChannelInteractive, "req-中", "run-a", "stream-b", TypeData, 4, []byte{0, 1, 2, 255}, 0}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	dynamic := dynamicpb.NewMessage(fd.Messages().Get(0))
	if err = proto.Unmarshal(encoded, dynamic); err != nil {
		t.Fatal(err)
	}
	if dynamic.Get(dynamic.Descriptor().Fields().ByNumber(4)).String() != original.RequestID || !bytes.Equal(dynamic.Get(dynamic.Descriptor().Fields().ByNumber(9)).Bytes(), original.Payload) {
		t.Fatal("protobuf decoded wrong payload")
	}
	dynamic.Set(dynamic.Descriptor().Fields().ByNumber(2), protoreflect.ValueOfUint64(24))
	encoded, err = proto.Marshal(dynamic)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := Unmarshal(encoded)
	if err != nil {
		t.Fatal(err)
	}
	original.Generation = 24
	if !reflect.DeepEqual(original, roundtrip) {
		t.Fatalf("roundtrip: %#v", roundtrip)
	}
	unknown := protowire.AppendTag(encoded, 99, protowire.BytesType)
	unknown = protowire.AppendString(unknown, "future-field")
	if _, err = Unmarshal(unknown); err != nil {
		t.Fatalf("compatible unknown field: %v", err)
	}
}

func TestMalformedFrames(t *testing.T) {
	valid, _ := Marshal(Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelControl, Type: TypePing})
	duplicate := append(append([]byte(nil), valid...), 0x10, 0x02)
	overflow := append([]byte{0x08}, protowire.AppendVarint(nil, 1<<40)...)
	for name, data := range map[string][]byte{"empty": {}, "truncated": {0x08}, "zero-tag": {0}, "duplicate": duplicate, "overflow": overflow, "group": {0x5b}, "oversize": make([]byte, MaxMessageSize+1), "string-wire": {0x0a, 1, 1}} {
		t.Run(name, func(t *testing.T) {
			if _, err := Unmarshal(data); err == nil {
				t.Fatal("accepted malformed frame")
			}
		})
	}
}

func TestFlowCreditAndReplay(t *testing.T) {
	f := NewFlowLedger(2, 8, 16)
	if err := f.Open("s", 8, 8, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := f.Open("other", 1, 0, 0, 0); !errors.Is(err, ErrNoCredit) {
		t.Fatalf("global bound: %v", err)
	}
	if err := f.ReserveSend("s", 8, 8); err != nil {
		t.Fatal(err)
	}
	if err := f.ReserveSend("s", 9, 1); !errors.Is(err, ErrNoCredit) {
		t.Fatalf("unconsumed send: %v", err)
	}
	if err := f.ReceiveAck("s", 9, 9); !errors.Is(err, ErrSequence) {
		t.Fatalf("future ack: %v", err)
	}
	if err := f.ReceiveAck("s", 8, 8); err != nil {
		t.Fatal(err)
	}
	if err := f.ReceiveAck("s", 8, 8); !errors.Is(err, ErrSequence) {
		t.Fatalf("duplicate inflated credit: %v", err)
	}
	if err := f.ReserveSend("s", 16, 8); err != nil {
		t.Fatal(err)
	}
	if err := f.ReceiveData("s", 8, 8); err != nil {
		t.Fatal(err)
	}
	if err := f.ReceiveData("s", 9, 1); !errors.Is(err, ErrNoCredit) {
		t.Fatalf("receive overflow: %v", err)
	}
	credit, err := f.Consumed("s", 4)
	if err != nil || credit != 4 {
		t.Fatalf("partial consumption: %d %v", credit, err)
	}
	if err = f.ReceiveData("s", 12, 4); err != nil {
		t.Fatal(err)
	}
	if err = f.ReceiveData("s", 12, 4); !errors.Is(err, ErrSequence) {
		t.Fatalf("replayed data: %v", err)
	}
	f.Close("s")
	if got := f.Stats(); got.Streams != 0 || got.AllocatedWindow != 0 {
		t.Fatalf("leaked stream budget: %+v", got)
	}
}

func TestChallengeBindsIdentityConnectionAndExpiry(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewChallenge("node-a", "connection-a", 2, ChannelControl, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := SignChallenge(priv, c)
	if err != nil {
		t.Fatal(err)
	}
	if err = VerifyChallenge(pub, c, signature); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*Challenge){func(c *Challenge) { c.NodeID = "node-b" }, func(c *Challenge) { c.ConnectionID = "connection-b" }, func(c *Challenge) { c.Generation++ }, func(c *Challenge) { c.Channel = ChannelBulk }, func(c *Challenge) { c.Nonce = strings.Repeat("A", 43) }, func(c *Challenge) { c.ExpiresAt = time.Now().Add(-time.Second).Unix() }, func(c *Challenge) { c.ProtocolVersion++ }}
	for i, mutate := range mutations {
		changed := c
		mutate(&changed)
		if err = VerifyChallenge(pub, changed, signature); err == nil {
			t.Fatalf("accepted changed challenge %d", i)
		}
	}
	if _, err = SignChallenge(nil, c); err == nil {
		t.Fatal("accepted absent private key")
	}
}

func wssPair(t *testing.T, options Options) (*Conn, *websocket.Conn) {
	t.Helper()
	accepted := make(chan *Conn, 1)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		c, err := NewConn(ctx, ws, options)
		if err != nil {
			t.Errorf("new conn: %v", err)
			ws.CloseNow()
			return
		}
		accepted <- c
	}))
	t.Cleanup(server.Close)
	dialCtx, dialCancel := context.WithTimeout(ctx, 3*time.Second)
	defer dialCancel()
	client, _, err := websocket.Dial(dialCtx, "wss"+strings.TrimPrefix(server.URL, "https"), &websocket.DialOptions{HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.CloseNow() })
	select {
	case c := <-accepted:
		t.Cleanup(func() { c.Close() })
		return c, client
	case <-dialCtx.Done():
		t.Fatal(dialCtx.Err())
		return nil, nil
	}
}

func writeEnvelope(t *testing.T, ws *websocket.Conn, e Envelope) {
	t.Helper()
	b, err := Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = ws.Write(ctx, websocket.MessageBinary, b); err != nil {
		t.Fatal(err)
	}
}

func TestWSSReadWriteAndGenerationFence(t *testing.T) {
	var generation atomic.Uint64
	generation.Store(3)
	server, client := wssPair(t, Options{Generation: 3, Channel: ChannelControl, CurrentGeneration: generation.Load})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	writeEnvelope(t, client, Envelope{ProtocolVersion: Version, Generation: 3, Channel: ChannelControl, Type: TypePing, RequestID: "ping"})
	got, err := server.Read(ctx)
	if err != nil || got.RequestID != "ping" {
		t.Fatalf("read: %+v %v", got, err)
	}
	if err = server.Send(ctx, Envelope{Type: TypePong, RequestID: got.RequestID}); err != nil {
		t.Fatal(err)
	}
	typ, b, err := client.Read(ctx)
	if err != nil || typ != websocket.MessageBinary {
		t.Fatalf("client read: %v", err)
	}
	reply, err := Unmarshal(b)
	if err != nil || reply.Type != TypePong || reply.Generation != 3 {
		t.Fatalf("reply: %+v %v", reply, err)
	}
	generation.Store(4)
	if err = server.Send(ctx, Envelope{Type: TypeTaskEvent}); !errors.Is(err, ErrGeneration) {
		t.Fatalf("old sender generation accepted: %v", err)
	}
}

func TestWSSFlowAcknowledgesConsumptionOnly(t *testing.T) {
	server, rawClient := wssPair(t, Options{Generation: 1, Channel: ChannelInteractive})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := NewConn(ctx, rawClient, Options{Generation: 1, Channel: ChannelInteractive})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, c := range []*Conn{server, client} {
		if err = c.Flow().Open("out", 4, 4, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err = server.Send(ctx, Envelope{Type: TypeData, StreamID: "out", Sequence: 4, Payload: []byte("test")}); err != nil {
		t.Fatal(err)
	}
	got, err := client.Read(ctx)
	if err != nil || string(got.Payload) != "test" {
		t.Fatalf("receive: %+v %v", got, err)
	}
	if err = server.Send(ctx, Envelope{Type: TypeData, StreamID: "out", Sequence: 5, Payload: []byte("!")}); !errors.Is(err, ErrNoCredit) {
		t.Fatalf("socket read incorrectly replenished: %v", err)
	}
	if err = client.Consume(ctx, "out", 4); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for server.Flow().Stats().SentUnacknowledged != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err = server.Send(ctx, Envelope{Type: TypeData, StreamID: "out", Sequence: 5, Payload: []byte("!")}); err != nil {
		t.Fatal(err)
	}
}

func TestWSSRejectsCrossChannelStaleMalformedAndOverflow(t *testing.T) {
	for _, kind := range []string{"channel", "generation", "text", "oversize", "receive-credit"} {
		t.Run(kind, func(t *testing.T) {
			channel := ChannelControl
			if kind == "receive-credit" {
				channel = ChannelInteractive
			}
			server, client := wssPair(t, Options{Generation: 5, Channel: channel})
			e := Envelope{ProtocolVersion: Version, Generation: 5, Channel: channel, Type: TypePing}
			switch kind {
			case "channel":
				e.Channel = ChannelBulk
			case "generation":
				e.Generation = 4
			case "receive-credit":
				if err := server.Flow().Open("bad", 1, 1, 0, 0); err != nil {
					t.Fatal(err)
				}
				e.Type = TypeData
				e.StreamID = "bad"
				e.Sequence = 2
				e.Payload = []byte("xx")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if kind == "text" {
				_ = client.Write(ctx, websocket.MessageText, []byte("{}"))
			} else if kind == "oversize" {
				_ = client.Write(ctx, websocket.MessageBinary, make([]byte, MaxControlMessageSize+1))
			} else {
				writeEnvelope(t, client, e)
			}
			if _, err := server.Read(ctx); err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("invalid connection not rejected: %v", err)
			}
		})
	}
}

func TestControlBurstWaitsForSlowConsumer(t *testing.T) {
	server, client := wssPair(t, Options{Generation: 1, Channel: ChannelControl, QueueSize: 2, QueueBytes: 4096})
	for i := 0; i < 20; i++ {
		writeEnvelope(t, client, Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelControl, Type: TypeOpen, RequestID: fmt.Sprint(i)})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for server.Stats().IncomingBytes < 3*128 {
		select {
		case <-server.Done():
			t.Fatal("control burst disconnected before consumer drained")
		case <-ctx.Done():
			t.Fatal("control queue did not fill")
		case <-time.After(time.Millisecond):
		}
	}
	stats := server.Stats()
	if stats.IncomingQueued != 2 || stats.IncomingBytes > 4096 {
		t.Fatalf("unbounded control queue: %+v", stats)
	}
	for i := 0; i < 20; i++ {
		e, err := server.Read(ctx)
		if err != nil || e.RequestID != fmt.Sprint(i) {
			t.Fatalf("lost or reordered command %d: %q %v", i, e.RequestID, err)
		}
	}
}

func TestSeparateBulkCannotStarveControl(t *testing.T) {
	bulk, bulkClient := wssPair(t, Options{Generation: 1, Channel: ChannelBulk, QueueSize: 2, QueueBytes: 512})
	control, controlClient := wssPair(t, Options{Generation: 1, Channel: ChannelControl})
	for i := 0; i < 3; i++ {
		writeEnvelope(t, bulkClient, Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelBulk, Type: TypeOpen, RequestID: fmt.Sprint(i), Payload: make([]byte, 100)})
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-bulk.Done():
	case <-ctx.Done():
		t.Fatal("unconsumed bulk queue did not enforce bound")
	}
	if bulk.Stats().IncomingBytes > 512 {
		t.Fatal("incoming memory exceeds configured bound")
	}
	writeEnvelope(t, controlClient, Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelControl, Type: TypeOpen, RequestID: "stop"})
	e, err := control.Read(ctx)
	if err != nil || e.RequestID != "stop" {
		t.Fatalf("bulk starved control: %v", err)
	}
	if err = control.Send(ctx, Envelope{Type: TypeOpenAck, RequestID: "stop"}); err != nil {
		t.Fatal(err)
	}
}

func TestCursorRequiresSnapshotAfterGap(t *testing.T) {
	var c Cursor
	if _, err := c.Event(1); !errors.Is(err, ErrGap) {
		t.Fatal("accepted event before snapshot")
	}
	c.Snapshot(8)
	if fresh, err := c.Event(8); err != nil || fresh {
		t.Fatal("duplicate event")
	}
	if fresh, err := c.Event(9); err != nil || !fresh {
		t.Fatal("next event")
	}
	if _, err := c.Event(11); !errors.Is(err, ErrGap) {
		t.Fatal("missed gap")
	}
	if _, err := c.Event(10); !errors.Is(err, ErrGap) {
		t.Fatal("silently recovered partial event state")
	}
	c.Snapshot(11)
	if fresh, err := c.Event(12); err != nil || !fresh {
		t.Fatal("snapshot recovery")
	}
	if err := ResumeAvailable(2, 3, 8); !errors.Is(err, ErrGap) {
		t.Fatal("invented removed history")
	}
}

func TestCancelledReadKeepsSessionAndCancelledSendDoesNotSubmit(t *testing.T) {
	server, client := wssPair(t, Options{Generation: 1, Channel: ChannelControl})
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := server.Read(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read: %v", err)
	}
	if err := server.Send(cancelled, Envelope{Type: TypeOpen, RequestID: "must-not-submit"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled send: %v", err)
	}
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	writeEnvelope(t, client, Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelControl, Type: TypePing, RequestID: "still-alive"})
	e, err := server.Read(ctx)
	if err != nil || e.RequestID != "still-alive" {
		t.Fatalf("cancelled local read closed session: %v", err)
	}
}

func TestBulkRateWaitHonorsWriteDeadline(t *testing.T) {
	server, client := wssPair(t, Options{Generation: 1, Channel: ChannelBulk, BulkBytesPerSecond: 1, WriteTimeout: 20 * time.Millisecond})
	client.SetReadLimit(MaxMessageSize)
	// The first two 600 KiB OPEN payloads exceed the bucket's 1 MiB burst.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Send(ctx, Envelope{Type: TypeOpen, Payload: make([]byte, 600*1024)}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.Read(ctx); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	err := server.Send(ctx, Envelope{Type: TypeOpen, Payload: make([]byte, 600*1024)})
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 500*time.Millisecond {
		t.Fatalf("bulk rate wait has no deadline: %v", err)
	}
}

func FuzzUnmarshal(f *testing.F) {
	b, _ := Marshal(Envelope{ProtocolVersion: Version, Generation: 1, Channel: ChannelControl, Type: TypePing})
	f.Add(b)
	f.Add([]byte{0xff, 0, 0x1a})
	f.Fuzz(func(t *testing.T, b []byte) {
		e, err := Unmarshal(b)
		if err == nil {
			encoded, err := Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Unmarshal(encoded); err != nil {
				t.Fatal(err)
			}
		}
	})
}
