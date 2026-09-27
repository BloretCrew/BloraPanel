package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestTransferProofUnalignedBytesAndRestoredMtime(t *testing.T) {
	s, root, _ := testService(t, Options{ChunkBytes: 256 << 10})
	data := make([]byte, 3*proofBlockBytes+17)
	for i := range data {
		data[i] = byte(i % 251)
	}
	p := filepath.Join(root, "source")
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	v := version(t, s, "source")
	first, err := s.TransferStat(ctx, "source", v)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := s.ReadTransferChunk(ctx, "source", v, proofBlockBytes-6, proofBlockBytes+20)
	if err != nil || !bytes.Equal(chunk.Data, data[proofBlockBytes-6:2*proofBlockBytes+14]) || chunk.Version != v || chunk.Hash != hashBytes(chunk.Data) {
		t.Fatalf("unaligned read: %v", err)
	}
	if len(s.proofs) != 1 || s.proofBytes != 4*32 {
		t.Fatalf("unexpected proof budget: %d/%d", len(s.proofs), s.proofBytes)
	}
	f, err := os.OpenFile(p, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteAt([]byte{255}, 2*proofBlockBytes+5)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	if err = os.Chtimes(p, first.Modified, first.Modified); err != nil {
		t.Fatal(err)
	}
	// A metadata-only version cache would silently accept this changed block.
	if _, err = s.ReadTransferChunk(ctx, "source", v, 2*proofBlockBytes, 20); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed block accepted: %v", err)
	}
	// Other old-version bytes remain provable, but the mandatory final Stat
	// detects the whole source change before Master publishes the destination.
	unchanged, err := s.ReadTransferChunk(ctx, "source", v, 0, 20)
	if err != nil || !bytes.Equal(unchanged.Data, data[:20]) {
		t.Fatalf("unmodified version bytes: %v", err)
	}
	current, err := s.Stat(ctx, "source")
	if err != nil || current.Version == v {
		t.Fatalf("final whole-file validation missed change: %v", err)
	}
}

func TestTransferProofEntryBudgetAndReplacement(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	for i := 0; i < proofMaxEntries+3; i++ {
		p := fmt.Sprintf("empty-%d", i)
		if err := os.WriteFile(filepath.Join(root, p), nil, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := s.TransferStat(ctx, p, hashBytes(nil)); err != nil {
			t.Fatal(err)
		}
	}
	if len(s.proofs) != proofMaxEntries || s.proofBytes > proofBudgetBytes {
		t.Fatalf("unbounded empty entries: %d/%d", len(s.proofs), s.proofBytes)
	}
	put(t, root, "source", "original")
	v := version(t, s, "source")
	first, err := s.TransferStat(ctx, "source", v)
	if err != nil {
		t.Fatal(err)
	}
	old := s.proofs["source\x00"+v]
	if err = os.Rename(filepath.Join(root, "source"), filepath.Join(root, "old")); err != nil {
		t.Fatal(err)
	}
	put(t, root, "source", "original")
	if err = os.Chtimes(filepath.Join(root, "source"), first.Modified, first.Modified); err != nil {
		t.Fatal(err)
	}
	if _, err = s.TransferStat(ctx, "source", v); err != nil {
		t.Fatal(err)
	}
	if old == s.proofs["source\x00"+v] {
		t.Fatal("replacement reused the old object's proof")
	}
	if _, err = s.ReadTransferChunk(ctx, "source", hashBytes([]byte("unknown")), 0, 1); !errors.Is(err, ErrConflict) {
		t.Fatalf("unverified version accepted: %v", err)
	}
}

func TestTransferProofConcurrentReaders(t *testing.T) {
	s, root, _ := testService(t, Options{})
	data := bytes.Repeat([]byte("bounded-proof"), 20000)
	if err := os.WriteFile(filepath.Join(root, "source"), data, 0600); err != nil {
		t.Fatal(err)
	}
	v := hashBytes(data)
	var workers sync.WaitGroup
	for i := 0; i < 24; i++ {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			offset := index * 1003
			chunk, err := s.ReadTransferChunk(context.Background(), "source", v, int64(offset), 65536)
			if err != nil || !bytes.Equal(chunk.Data, data[offset:offset+65536]) {
				t.Errorf("reader %d: %v", index, err)
			}
		}(i)
	}
	workers.Wait()
	if len(s.proofs) != 1 || s.proofBytes > proofBudgetBytes {
		t.Fatal("concurrent fills escaped cache accounting")
	}
}
