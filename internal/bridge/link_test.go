package bridge

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"github.com/coder/websocket"
)

func bridgePair(t *testing.T) (context.Context, *protocol.Conn, *protocol.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	accepted := make(chan *protocol.Conn, 1)
	options := protocol.Options{Generation: 1, Channel: protocol.ChannelBulk, InitialStreams: []string{StreamID}}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		conn, err := protocol.NewConn(ctx, ws, options)
		if err != nil {
			t.Error(err)
			_ = ws.CloseNow()
			return
		}
		accepted <- conn
		<-conn.Done()
	}))
	t.Cleanup(server.Close)
	ws, _, err := websocket.Dial(ctx, "wss"+strings.TrimPrefix(server.URL, "https"), &websocket.DialOptions{HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	client, err := protocol.NewConn(ctx, ws, options)
	if err != nil {
		_ = ws.CloseNow()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	select {
	case peer := <-accepted:
		t.Cleanup(func() { _ = peer.Close() })
		return ctx, peer, client
	case <-ctx.Done():
		t.Fatal(ctx.Err())
		return nil, nil, nil
	}
}

func TestSendPreservesClosedTransportIdentity(t *testing.T) {
	ctx, peer, client := bridgePair(t)
	// Deliberately leave Link.read() absent to exercise the interval before the
	// bridge reader has translated a raw socket failure into Link cancellation.
	_ = client.Close()
	select {
	case <-peer.Done():
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	cause := peer.Err()
	if cause == nil || errors.Is(cause, protocol.ErrClosed) {
		t.Fatalf("fixture did not produce an underlying socket failure: %v", cause)
	}
	link := &Link{conn: peer, ctx: ctx}
	err := link.send(ctx, frame{Request: &Request{ID: model.ID(), Method: "file.read", ActorID: "owner"}})
	if !errors.Is(err, protocol.ErrClosed) || !errors.Is(err, cause) || link.sent != 0 {
		t.Fatalf("closed transport identity or cause lost: %v, sent=%d", err, link.sent)
	}
}

func TestRemoteOperationErrorIsNotTransportLoss(t *testing.T) {
	ctx, peer, client := bridgePair(t)
	remote := New(ctx, peer, func(context.Context, Request) (any, error) { return nil, io.EOF })
	local := New(ctx, client, nil)
	t.Cleanup(func() { _ = local.Close(); _ = remote.Close(); local.Wait(); remote.Wait() })
	err := local.Call(ctx, Request{Method: "file.read", ActorID: "owner"}, nil)
	var api *model.APIError
	if !errors.As(err, &api) || api.Code != "NODE_OPERATION_FAILED" || api.Message != io.EOF.Error() || errors.Is(err, protocol.ErrClosed) {
		t.Fatalf("remote file failure mistaken for reconnectable transport loss: %v", err)
	}
}
