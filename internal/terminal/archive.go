// Package terminal manages background terminals and bounded, byte-preserving
// output archives. Authorization and transport byte credits are separate from
// the archive's monotonically increasing event cursor.
package terminal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const archiveHeader = 32
const MaxEventBytes = 32 << 10
const MaxBatchBytes = 1 << 20

var ErrCursor = errors.New("terminal cursor is ahead of the archive")
var ErrArchive = errors.New("terminal archive is corrupt")

type ArchiveOptions struct {
	MaxBytes     int64 `json:"maxBytes"`
	SegmentBytes int64 `json:"segmentBytes"`
}

type Event struct {
	Sequence uint64 `json:"sequence"`
	Kind     string `json:"kind"`
	Data     []byte `json:"data,omitempty"`
	Cols     uint16 `json:"cols,omitempty"`
	Rows     uint16 `json:"rows,omitempty"`
}

type Batch struct {
	Events   []Event `json:"events"`
	Earliest uint64  `json:"earliest"`
	Latest   uint64  `json:"latest"`
	Next     uint64  `json:"next"`
	Gap      bool    `json:"gap"`
}

type segment struct {
	name        string
	first, last uint64
	size        int64
}
type Archive struct {
	mu       sync.Mutex
	root     string
	options  ArchiveOptions
	segments []segment
	bytes    int64
	latest   uint64
	closed   bool
}

func archiveOptions(o ArchiveOptions) (ArchiveOptions, error) {
	if o.MaxBytes == 0 {
		o.MaxBytes = 16 << 20
	}
	if o.SegmentBytes == 0 {
		o.SegmentBytes = min(1<<20, o.MaxBytes/2)
	}
	if o.SegmentBytes < 128 || o.MaxBytes < 2*o.SegmentBytes || o.MaxBytes > 1<<30 {
		return o, errors.New("archive requires 128 <= segmentBytes <= maxBytes/2 and maxBytes <= 1 GiB")
	}
	return o, nil
}

// OpenArchive recovers complete records, truncating only an incomplete last
// record. Full-record corruption fails closed. One Manager must own a root.
func OpenArchive(root string, options ArchiveOptions) (*Archive, error) {
	o, err := archiveOptions(options)
	if err != nil {
		return nil, err
	}
	if root == "" {
		return nil, errors.New("archive root is required")
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	a := &Archive{root: root, options: o}
	var names []string
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".seg") {
			continue
		}
		if e.IsDir() || e.Type()&os.ModeSymlink != 0 {
			return nil, ErrArchive
		}
		id, err := strconv.ParseUint(strings.TrimSuffix(e.Name(), ".seg"), 10, 64)
		if err != nil || id == 0 || e.Name() != fmt.Sprintf("%020d.seg", id) {
			return nil, ErrArchive
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for i, name := range names {
		f, err := os.OpenFile(filepath.Join(root, name), os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		s := segment{name: name}
		for {
			e, size, err := readEvent(f)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				if errors.Is(err, io.ErrUnexpectedEOF) && i == len(names)-1 {
					err = f.Truncate(s.size)
					if err == nil {
						err = f.Sync()
					}
					if err != nil {
						f.Close()
						return nil, err
					}
					break
				}
				f.Close()
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			if (a.latest != 0 && e.Sequence != a.latest+1) || (s.first == 0 && name != fmt.Sprintf("%020d.seg", e.Sequence)) {
				f.Close()
				return nil, ErrArchive
			}
			if s.first == 0 {
				s.first = e.Sequence
			}
			s.last = e.Sequence
			s.size += int64(size)
			a.latest = e.Sequence
		}
		if err = f.Close(); err != nil {
			return nil, err
		}
		if s.size == 0 {
			if err = os.Remove(filepath.Join(root, name)); err != nil {
				return nil, err
			}
			continue
		}
		a.segments = append(a.segments, s)
		a.bytes += s.size
	}
	if err = a.prune(0); err != nil {
		return nil, err
	}
	return a, nil
}

func readEvent(r io.Reader) (Event, int, error) {
	var h [archiveHeader]byte
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return Event{}, 0, err
	}
	if string(h[:4]) != "BTA1" {
		return Event{}, 0, ErrArchive
	}
	n := binary.LittleEndian.Uint32(h[20:24])
	if n > MaxEventBytes {
		return Event{}, 0, ErrArchive
	}
	e := Event{Sequence: binary.LittleEndian.Uint64(h[4:12]), Cols: binary.LittleEndian.Uint16(h[14:16]), Rows: binary.LittleEndian.Uint16(h[16:18])}
	switch h[12] {
	case 1:
		e.Kind = "output"
		if n == 0 || e.Cols != 0 || e.Rows != 0 {
			return Event{}, 0, ErrArchive
		}
	case 2:
		e.Kind = "resize"
		if n != 0 || !validSize(e.Cols, e.Rows) {
			return Event{}, 0, ErrArchive
		}
	default:
		return Event{}, 0, ErrArchive
	}
	if e.Sequence == 0 {
		return Event{}, 0, ErrArchive
	}
	e.Data = make([]byte, n)
	if _, err := io.ReadFull(r, e.Data); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return Event{}, 0, err
	}
	checksum := crc32.NewIEEE()
	checksum.Write(h[:24])
	checksum.Write(e.Data)
	if checksum.Sum32() != binary.LittleEndian.Uint32(h[24:28]) {
		return Event{}, 0, ErrArchive
	}
	return e, archiveHeader + int(n), nil
}

func (a *Archive) prune(extra int64) error {
	changed := false
	for a.bytes+extra > a.options.MaxBytes && len(a.segments) > 1 {
		s := a.segments[0]
		if err := os.Remove(filepath.Join(a.root, s.name)); err != nil {
			return err
		}
		a.segments = a.segments[1:]
		a.bytes -= s.size
		changed = true
	}
	if a.bytes+extra > a.options.MaxBytes {
		return errors.New("archive budget cannot retain current segment")
	}
	if changed {
		return syncDirectory(a.root)
	}
	return nil
}

func (a *Archive) append(e Event) error {
	if a.closed {
		return os.ErrClosed
	}
	e.Sequence = a.latest + 1
	if e.Sequence == 0 {
		return errors.New("archive cursor exhausted")
	}
	size := int64(archiveHeader + len(e.Data))
	if len(a.segments) == 0 || a.segments[len(a.segments)-1].size+size > a.options.SegmentBytes {
		a.segments = append(a.segments, segment{name: fmt.Sprintf("%020d.seg", e.Sequence), first: e.Sequence})
	}
	if err := a.prune(size); err != nil {
		return err
	}
	s := &a.segments[len(a.segments)-1]
	newSegment := s.size == 0
	f, err := os.OpenFile(filepath.Join(a.root, s.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	var h [archiveHeader]byte
	copy(h[:4], "BTA1")
	binary.LittleEndian.PutUint64(h[4:12], e.Sequence)
	if e.Kind == "output" {
		h[12] = 1
	} else {
		h[12] = 2
	}
	binary.LittleEndian.PutUint16(h[14:16], e.Cols)
	binary.LittleEndian.PutUint16(h[16:18], e.Rows)
	binary.LittleEndian.PutUint32(h[20:24], uint32(len(e.Data)))
	checksum := crc32.NewIEEE()
	checksum.Write(h[:24])
	checksum.Write(e.Data)
	binary.LittleEndian.PutUint32(h[24:28], checksum.Sum32())
	record := append(h[:], e.Data...)
	n, err := f.Write(record)
	if err == nil && n != len(record) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		_ = f.Truncate(s.size)
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	s.last = e.Sequence
	s.size += size
	a.bytes += size
	a.latest = e.Sequence
	if newSegment {
		return syncDirectory(a.root)
	}
	return nil
}

// AppendOutput preserves raw bytes, including UTF-8 fragments and ANSI escapes.
// It never stores input. Large writes are split into bounded ordered records.
func (a *Archive) AppendOutput(data []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	chunk := min(MaxEventBytes, int(a.options.SegmentBytes)-archiveHeader)
	for len(data) > 0 {
		n := min(chunk, len(data))
		if err := a.append(Event{Kind: "output", Data: data[:n]}); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
func (a *Archive) AppendResize(cols, rows uint16) error {
	if !validSize(cols, rows) {
		return errors.New("invalid terminal size")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.append(Event{Kind: "resize", Cols: cols, Rows: rows})
}
func (a *Archive) bounds() (uint64, uint64) {
	if len(a.segments) == 0 {
		return 1, a.latest
	}
	return a.segments[0].first, a.latest
}
func (a *Archive) Bounds() (uint64, uint64) { a.mu.Lock(); defer a.mu.Unlock(); return a.bounds() }
func (a *Archive) Bytes() int64             { a.mu.Lock(); defer a.mu.Unlock(); return a.bytes }

// Read uses an exclusive event cursor. The effective batch budget is clamped
// to [MaxEventBytes+32, MaxBatchBytes], counting headers as well as raw output.
// Gap means the checkpoint's next event was already retired; never silently
// apply these events to that checkpoint. No browser-dependent queue is kept.
func (a *Archive) Read(after uint64, maxBytes int) (Batch, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return Batch{}, os.ErrClosed
	}
	earliest, latest := a.bounds()
	b := Batch{Events: []Event{}, Earliest: earliest, Latest: latest, Next: after}
	if after > latest {
		return b, ErrCursor
	}
	if after < earliest-1 {
		b.Gap = true
		b.Next = earliest - 1
	}
	if maxBytes <= 0 {
		maxBytes = 64 << 10
	}
	budget := max(MaxEventBytes+archiveHeader, min(maxBytes, MaxBatchBytes))
	used := 0
	for _, s := range a.segments {
		if s.last <= b.Next {
			continue
		}
		f, err := os.Open(filepath.Join(a.root, s.name))
		if err != nil {
			return b, err
		}
		for {
			e, size, err := readEvent(f)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				f.Close()
				return b, err
			}
			if e.Sequence <= b.Next {
				continue
			}
			if used+size > budget {
				f.Close()
				return b, nil
			}
			b.Events = append(b.Events, e)
			b.Next = e.Sequence
			used += size
		}
		if err = f.Close(); err != nil {
			return b, err
		}
	}
	return b, nil
}
func (a *Archive) Close() error { a.mu.Lock(); defer a.mu.Unlock(); a.closed = true; return nil }

func validSize(cols, rows uint16) bool { return cols > 0 && rows > 0 && cols <= 1000 && rows <= 1000 }
