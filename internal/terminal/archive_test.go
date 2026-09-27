package terminal

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveRawBytesResizeOrderingRecovery(t *testing.T) {
	root := t.TempDir()
	o := ArchiveOptions{MaxBytes: 1 << 20, SegmentBytes: 4096}
	a, err := OpenArchive(root, o)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("\x1b[?1049h\x1b[31m中文\x1b[0m\x00\xff")
	if err = a.AppendOutput(raw[:12]); err != nil {
		t.Fatal(err)
	}
	if err = a.AppendResize(132, 43); err != nil {
		t.Fatal(err)
	}
	if err = a.AppendOutput(raw[12:]); err != nil {
		t.Fatal(err)
	}
	b, err := a.Read(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if b.Gap || b.Earliest != 1 || b.Latest != 3 || b.Next != 3 || len(b.Events) != 3 {
		t.Fatalf("invalid bounds: %+v", b)
	}
	if b.Events[1].Kind != "resize" || b.Events[1].Cols != 132 || b.Events[1].Rows != 43 {
		t.Fatal("resize not ordered")
	}
	if !bytes.Equal(append(b.Events[0].Data, b.Events[2].Data...), raw) {
		t.Fatal("raw UTF-8/ANSI bytes changed")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	a, err = OpenArchive(root, o)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err = a.Read(1, 0)
	if err != nil || b.Next != 3 || len(b.Events) != 2 {
		t.Fatalf("exclusive resume: %+v %v", b, err)
	}
	if _, err = a.Read(4, 0); !errors.Is(err, ErrCursor) {
		t.Fatal("ahead cursor accepted")
	}
}

func TestArchiveRetentionBudgetAndGap(t *testing.T) {
	root := t.TempDir()
	a, err := OpenArchive(root, ArchiveOptions{MaxBytes: 1024, SegmentBytes: 256})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if err = a.AppendOutput(bytes.Repeat([]byte{byte(i)}, 224)); err != nil {
			t.Fatal(err)
		}
		if a.Bytes() > 1024 {
			t.Fatal("archive exceeds disk budget")
		}
	}
	b, err := a.Read(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !b.Gap || b.Earliest <= 1 || b.Latest != 100 || b.Next != 100 {
		t.Fatalf("retention must expose gap: %+v", b)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var size int64
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		size += info.Size()
	}
	if size > 1024 || size != a.Bytes() {
		t.Fatalf("disk bytes=%d tracking=%d", size, a.Bytes())
	}
	a.Close()
	a, err = OpenArchive(root, ArchiveOptions{MaxBytes: 1024, SegmentBytes: 256})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.AppendResize(80, 24); err != nil {
		t.Fatal(err)
	}
	_, latest := a.Bounds()
	if latest != 101 {
		t.Fatalf("cursor reset after rotation: %d", latest)
	}
}

func TestArchivePartialTailAndCorruption(t *testing.T) {
	root := t.TempDir()
	a, err := OpenArchive(root, ArchiveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.AppendOutput([]byte("complete")); err != nil {
		t.Fatal(err)
	}
	a.Close()
	name := filepath.Join(root, "00000000000000000001.seg")
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("BTA1partial")); err != nil {
		t.Fatal(err)
	}
	f.Close()
	a, err = OpenArchive(root, ArchiveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := a.Read(0, 0)
	if err != nil || b.Latest != 1 {
		t.Fatalf("partial recovery: %+v %v", b, err)
	}
	a.Close()
	f, err = os.OpenFile(name, os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteAt([]byte("x"), archiveHeader); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err = OpenArchive(root, ArchiveOptions{}); !errors.Is(err, ErrArchive) {
		t.Fatalf("corruption not rejected: %v", err)
	}
}

func TestArchiveLargeWriteAndResizeFloodBounded(t *testing.T) {
	a, err := OpenArchive(t.TempDir(), ArchiveOptions{MaxBytes: 2 << 20, SegmentBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.AppendOutput(bytes.Repeat([]byte("x"), 2*MaxEventBytes+7)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1200; i++ {
		if err = a.AppendResize(80, 24); err != nil {
			t.Fatal(err)
		}
	}
	b, err := a.Read(0, MaxEventBytes+archiveHeader)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Events) != 1 || len(b.Events[0].Data) != MaxEventBytes {
		t.Fatal("output record/batch limit ignored")
	}
	b, err = a.Read(3, MaxEventBytes+archiveHeader)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Events) > 1025 || b.Next >= b.Latest {
		t.Fatal("resize metadata bypassed batch budget")
	}
}
