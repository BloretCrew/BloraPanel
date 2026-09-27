package protocol

import "sync"

const DefaultStreamWindow uint64 = 256 * 1024
const DefaultConnectionWindow uint64 = 8 * 1024 * 1024
const DefaultMaxStreams = 64

type flowState struct {
	sendEnd, sendAck, sendLimit          uint64
	receiveEnd, receiveAck, receiveLimit uint64
	sendWindow, receiveWindow            uint64
}

// FlowLedger accounts bytes accepted and actually consumed, independently per
// stream and direction. It never grants credit from socket-write completion.
type FlowLedger struct {
	mu                             sync.Mutex
	streams                        map[string]*flowState
	maxStreams                     int
	maxWindow, maxTotal, allocated uint64
}

func NewFlowLedger(maxStreams int, maxWindow, maxTotal uint64) *FlowLedger {
	if maxStreams <= 0 {
		maxStreams = DefaultMaxStreams
	}
	if maxWindow == 0 {
		maxWindow = DefaultStreamWindow
	}
	if maxTotal == 0 {
		maxTotal = DefaultConnectionWindow
	}
	return &FlowLedger{streams: make(map[string]*flowState), maxStreams: maxStreams, maxWindow: maxWindow, maxTotal: maxTotal}
}

// Open registers windows negotiated by the authorized stream OPEN/OPEN_ACK.
// resumeSend/resumeReceive are acknowledged checkpoints, never historical input
// to resend. Reconnection requires the service to authorize/open a new ledger.
func (f *FlowLedger) Open(id string, sendWindow, receiveWindow, resumeSend, resumeReceive uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id == "" || len(id) > MaxIdentifierSize {
		return ErrStream
	}
	if _, ok := f.streams[id]; ok {
		return ErrStream
	}
	if len(f.streams) >= f.maxStreams {
		return ErrStreamLimit
	}
	if sendWindow > f.maxWindow || receiveWindow > f.maxWindow || sendWindow > f.maxTotal-f.allocated || receiveWindow > f.maxTotal-f.allocated-sendWindow {
		return ErrNoCredit
	}
	if resumeSend+sendWindow < resumeSend || resumeReceive+receiveWindow < resumeReceive {
		return ErrSequence
	}
	f.streams[id] = &flowState{sendEnd: resumeSend, sendAck: resumeSend, sendLimit: resumeSend + sendWindow, receiveEnd: resumeReceive, receiveAck: resumeReceive, receiveLimit: resumeReceive + receiveWindow, sendWindow: sendWindow, receiveWindow: receiveWindow}
	f.allocated += sendWindow + receiveWindow
	return nil
}

func (f *FlowLedger) Close(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.streams[id]; ok {
		f.allocated -= s.sendWindow + s.receiveWindow
		delete(f.streams, id)
	}
}

// ReserveSend provides immediate backpressure. The caller retains unsubmitted
// bytes or resumes from its bounded archive after consumers replenish credit.
func (f *FlowLedger) ReserveSend(id string, end, size uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.streams[id]
	if !ok {
		return ErrStream
	}
	if size == 0 || s.sendEnd+size < s.sendEnd || end != s.sendEnd+size {
		return ErrSequence
	}
	if end > s.sendLimit {
		return ErrNoCredit
	}
	s.sendEnd = end
	return nil
}

func (f *FlowLedger) refundSend(id string, end, size uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.streams[id]; ok && s.sendEnd == end {
		s.sendEnd -= size
	}
}

func (f *FlowLedger) ReceiveData(id string, end, size uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.streams[id]
	if !ok {
		return ErrStream
	}
	if size == 0 || s.receiveEnd+size < s.receiveEnd || end != s.receiveEnd+size {
		return ErrSequence
	}
	if end > s.receiveLimit {
		return ErrNoCredit
	}
	s.receiveEnd = end
	return nil
}

func (f *FlowLedger) ReceiveAck(id string, end, credit uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.streams[id]
	if !ok {
		return ErrStream
	}
	// A repeated identical ACK is harmless and must never replenish twice.
	if end == s.sendAck {
		if credit == 0 {
			return nil
		}
		return ErrSequence
	}
	if end < s.sendAck || end > s.sendEnd || credit != end-s.sendAck || s.sendLimit+credit < s.sendLimit {
		return ErrSequence
	}
	s.sendAck = end
	s.sendLimit += credit
	return nil
}

// ReceiveCredit accepts an absolute cumulative byte ceiling, making duplicate
// CREDIT frames idempotent. ACK is the normal way to return processed credit.
func (f *FlowLedger) ReceiveCredit(id string, ceiling uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.streams[id]
	if !ok {
		return ErrStream
	}
	if ceiling < s.sendLimit || ceiling < s.sendAck || ceiling-s.sendAck > s.sendWindow {
		return ErrNoCredit
	}
	s.sendLimit = ceiling
	return nil
}

// Consumed must only be called once the destination (e.g. xterm parse callback
// or durable upload checkpoint) has processed bytes through end.
func (f *FlowLedger) Consumed(id string, end uint64) (uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.streams[id]
	if !ok {
		return 0, ErrStream
	}
	if end < s.receiveAck || end > s.receiveEnd {
		return 0, ErrSequence
	}
	credit := end - s.receiveAck
	if s.receiveLimit+credit < s.receiveLimit {
		return 0, ErrSequence
	}
	s.receiveAck = end
	s.receiveLimit += credit
	return credit, nil
}

type FlowStats struct {
	Streams            int
	AllocatedWindow    uint64
	SentUnacknowledged uint64
	ReceivedUnconsumed uint64
}

func (f *FlowLedger) Stats() FlowStats {
	f.mu.Lock()
	defer f.mu.Unlock()
	v := FlowStats{Streams: len(f.streams), AllocatedWindow: f.allocated}
	for _, s := range f.streams {
		v.SentUnacknowledged += s.sendEnd - s.sendAck
		v.ReceivedUnconsumed += s.receiveEnd - s.receiveAck
	}
	return v
}
