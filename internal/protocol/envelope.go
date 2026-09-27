// Package protocol implements the bounded management protocol described in
// api/management.proto. Authorization and persistent request deduplication are
// responsibilities of the services using this transport.
package protocol

import (
	"errors"
	"fmt"
	"math"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"
)

const Version uint32 = 1
const MaxMessageSize = 1024 * 1024
const MaxControlMessageSize = 64 * 1024
const MaxIdentifierSize = 256

type Channel uint32

const (
	ChannelControl Channel = iota + 1
	ChannelInteractive
	ChannelBulk
)

func (c Channel) Valid() bool { return c >= ChannelControl && c <= ChannelBulk }
func (c Channel) String() string {
	switch c {
	case ChannelControl:
		return "control"
	case ChannelInteractive:
		return "interactive"
	case ChannelBulk:
		return "bulk"
	}
	return "unknown"
}
func ParseChannel(s string) (Channel, error) {
	switch s {
	case "control":
		return ChannelControl, nil
	case "interactive":
		return ChannelInteractive, nil
	case "bulk":
		return ChannelBulk, nil
	}
	return 0, ErrChannel
}

type Type uint32

const (
	TypeOpen Type = iota + 1
	TypeOpenAck
	TypeData
	TypeCredit
	TypeAck
	TypeResize
	TypeClose
	TypeReset
	TypeError
	TypeResume
	TypePing
	TypePong
	TypeTaskEvent
	TypeSnapshot
)

var (
	ErrMalformed       = errors.New("malformed management frame")
	ErrVersion         = errors.New("unsupported protocol version")
	ErrGeneration      = errors.New("stale connection generation")
	ErrChannel         = errors.New("message belongs to a different physical channel")
	ErrMessageTooLarge = errors.New("management message exceeds limit")
	ErrQueueFull       = errors.New("management queue is full")
	ErrClosed          = errors.New("management connection closed")
	ErrNoCredit        = errors.New("stream credit exhausted")
	ErrStream          = errors.New("unknown, closed, or duplicate stream")
	ErrSequence        = errors.New("invalid stream sequence or acknowledgement")
	ErrStreamLimit     = errors.New("active stream limit exceeded")
	ErrGap             = errors.New("event gap: refresh resource snapshot")
)

type Envelope struct {
	ProtocolVersion uint32  `json:"protocolVersion"`
	Generation      uint64  `json:"generation"`
	Channel         Channel `json:"channel"`
	RequestID       string  `json:"requestId,omitempty"`
	RunID           string  `json:"runId,omitempty"`
	StreamID        string  `json:"streamId,omitempty"`
	Type            Type    `json:"type"`
	Sequence        uint64  `json:"sequence,omitempty"`
	Payload         []byte  `json:"payload,omitempty"`
	Credit          uint64  `json:"credit,omitempty"`
}

func (e Envelope) Validate() error {
	if e.ProtocolVersion != Version {
		return ErrVersion
	}
	if e.Generation == 0 {
		return ErrGeneration
	}
	if !e.Channel.Valid() {
		return ErrChannel
	}
	if e.Type < TypeOpen || e.Type > TypeSnapshot {
		return fmt.Errorf("%w: message type", ErrMalformed)
	}
	for _, id := range []string{e.RequestID, e.RunID, e.StreamID} {
		if len(id) > MaxIdentifierSize || !utf8.ValidString(id) {
			return fmt.Errorf("%w: identifier", ErrMalformed)
		}
	}
	if e.Type == TypeData || e.Type == TypeAck || e.Type == TypeCredit || e.Type == TypeResize {
		if e.StreamID == "" {
			return ErrStream
		}
		if e.Channel == ChannelControl {
			return ErrChannel
		}
	}
	if e.Type == TypeData && len(e.Payload) == 0 {
		return fmt.Errorf("%w: empty DATA", ErrMalformed)
	}
	if len(e.Payload) > MaxMessageSize || (e.Channel == ChannelControl && len(e.Payload) > MaxControlMessageSize) {
		return ErrMessageTooLarge
	}
	return nil
}

func Marshal(e Envelope) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	b := make([]byte, 0, len(e.Payload)+128)
	uintField := func(n protowire.Number, v uint64) {
		if v != 0 {
			b = protowire.AppendTag(b, n, protowire.VarintType)
			b = protowire.AppendVarint(b, v)
		}
	}
	bytesField := func(n protowire.Number, v []byte) {
		if len(v) > 0 {
			b = protowire.AppendTag(b, n, protowire.BytesType)
			b = protowire.AppendBytes(b, v)
		}
	}
	uintField(1, uint64(e.ProtocolVersion))
	uintField(2, e.Generation)
	uintField(3, uint64(e.Channel))
	bytesField(4, []byte(e.RequestID))
	bytesField(5, []byte(e.RunID))
	bytesField(6, []byte(e.StreamID))
	uintField(7, uint64(e.Type))
	uintField(8, e.Sequence)
	bytesField(9, e.Payload)
	uintField(10, e.Credit)
	limit := MaxMessageSize
	if e.Channel == ChannelControl {
		limit = MaxControlMessageSize
	}
	if len(b) > limit {
		return nil, ErrMessageTooLarge
	}
	return b, nil
}

// Unmarshal accepts unknown non-group fields for compatible schema extensions,
// but rejects duplicate security-sensitive envelope fields and integer overflow.
func Unmarshal(b []byte) (Envelope, error) {
	var e Envelope
	if len(b) > MaxMessageSize {
		return e, ErrMessageTooLarge
	}
	size := len(b)
	var seen uint16
	for len(b) > 0 {
		n, typ, k := protowire.ConsumeTag(b)
		if k < 0 || !n.IsValid() || typ == protowire.StartGroupType || typ == protowire.EndGroupType {
			return e, ErrMalformed
		}
		b = b[k:]
		if n > 10 {
			k = protowire.ConsumeFieldValue(n, typ, b)
			if k < 0 {
				return e, ErrMalformed
			}
			b = b[k:]
			continue
		}
		if seen&(1<<n) != 0 {
			return e, fmt.Errorf("%w: duplicate field %d", ErrMalformed, n)
		}
		seen |= 1 << n
		if n == 4 || n == 5 || n == 6 || n == 9 {
			if typ != protowire.BytesType {
				return e, ErrMalformed
			}
			v, m := protowire.ConsumeBytes(b)
			if m < 0 {
				return e, ErrMalformed
			}
			b = b[m:]
			switch n {
			case 4:
				e.RequestID = string(v)
			case 5:
				e.RunID = string(v)
			case 6:
				e.StreamID = string(v)
			case 9:
				e.Payload = append([]byte(nil), v...)
			}
		} else {
			if typ != protowire.VarintType {
				return e, ErrMalformed
			}
			v, m := protowire.ConsumeVarint(b)
			if m < 0 {
				return e, ErrMalformed
			}
			b = b[m:]
			if (n == 1 || n == 3 || n == 7) && v > math.MaxUint32 {
				return e, ErrMalformed
			}
			switch n {
			case 1:
				e.ProtocolVersion = uint32(v)
			case 2:
				e.Generation = v
			case 3:
				e.Channel = Channel(v)
			case 7:
				e.Type = Type(v)
			case 8:
				e.Sequence = v
			case 10:
				e.Credit = v
			}
		}
	}
	if e.Channel == ChannelControl && size > MaxControlMessageSize {
		return e, ErrMessageTooLarge
	}
	return e, e.Validate()
}
