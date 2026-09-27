package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
)

const proofBlockBytes = 64 << 10
const proofBudgetBytes = 8 << 20
const proofMaxEntries = 64

// readProof binds every returned block to a whole-file digest verified in one
// pass. It contains hashes, not source data, and never survives a daemon restart.
// Metadata checks detect ordinary changes promptly; block hashes also detect
// writes which restore mtime. Transfer callers still recheck the whole source
// before publication and verify the complete destination digest.
type readProof struct {
	info   fs.FileInfo
	blocks [][sha256.Size]byte
	used   uint64
}

func sameProofFile(a, b fs.FileInfo) bool {
	return os.SameFile(a, b) && a.Size() == b.Size() && a.Mode() == b.Mode() && a.ModTime().Equal(b.ModTime())
}

func (s *Service) forgetProofLocked(key string) {
	if proof := s.proofs[key]; proof != nil {
		s.proofBytes -= len(proof.blocks) * sha256.Size
		delete(s.proofs, key)
	}
}

// Caller holds an ordinary read claim. Returned handle must be closed.
func (s *Service) openProven(ctx context.Context, p, version string) (*os.File, *readProof, error) {
	if !validHash(version) {
		return nil, nil, ErrConflict
	}
	f, info, err := s.openRegular(p)
	if err != nil {
		return nil, nil, err
	}
	ok := false
	defer func() {
		if !ok {
			f.Close()
		}
	}()
	if info.Size() < 0 || info.Size() > s.opts.MaxFileBytes {
		return nil, nil, ErrLimit
	}
	key := p + "\x00" + version
	s.proofMu.Lock()
	if proof := s.proofs[key]; proof != nil {
		if sameProofFile(proof.info, info) {
			s.proofClock++
			proof.used = s.proofClock
			s.proofMu.Unlock()
			ok = true
			return f, proof, nil
		}
		s.forgetProofLocked(key)
	}
	s.proofMu.Unlock()
	count := info.Size() / proofBlockBytes
	if info.Size()%proofBlockBytes != 0 {
		count++
	}
	if count > proofBudgetBytes/sha256.Size {
		return nil, nil, ErrLimit
	}
	proof := &readProof{info: info, blocks: make([][sha256.Size]byte, 0, int(count))}
	whole := sha256.New()
	buffer := make([]byte, proofBlockBytes)
	var total int64
	for {
		if err = ctx.Err(); err != nil {
			return nil, nil, err
		}
		n, readErr := io.ReadFull(f, buffer)
		if n > 0 {
			total += int64(n)
			if total > info.Size() {
				return nil, nil, ErrConflict
			}
			whole.Write(buffer[:n])
			proof.blocks = append(proof.blocks, sha256.Sum256(buffer[:n]))
		}
		if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}
		if readErr != nil {
			return nil, nil, readErr
		}
	}
	if total != info.Size() || "sha256:"+hex.EncodeToString(whole.Sum(nil)) != version {
		return nil, nil, ErrConflict
	}
	if err = s.checkProvenHandle(p, f, info); err != nil {
		return nil, nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, nil, err
	}
	s.proofMu.Lock()
	defer s.proofMu.Unlock()
	if s.proofs == nil {
		s.proofs = make(map[string]*readProof)
	}
	s.forgetProofLocked(key)
	bytes := len(proof.blocks) * sha256.Size
	for len(s.proofs) >= proofMaxEntries || s.proofBytes+bytes > proofBudgetBytes {
		var oldest string
		var age uint64
		for existing, value := range s.proofs {
			if oldest == "" || value.used < age {
				oldest = existing
				age = value.used
			}
		}
		s.forgetProofLocked(oldest)
	}
	s.proofClock++
	proof.used = s.proofClock
	s.proofs[key] = proof
	s.proofBytes += bytes
	ok = true
	return f, proof, nil
}

func (s *Service) checkProvenHandle(p string, f *os.File, initial fs.FileInfo) error {
	after, err := f.Stat()
	if err != nil {
		return err
	}
	current, err := s.root.Lstat(p)
	if err != nil {
		return err
	}
	if !sameProofFile(initial, after) || !sameProofFile(initial, current) {
		return ErrConflict
	}
	return nil
}

// TransferStat uses an in-memory proof for intermediate transfer checkpoints.
// This is not a replacement for Stat at submission or final source validation.
func (s *Service) TransferStat(ctx context.Context, p, version string) (Entry, error) {
	if err := s.validate(p, false); err != nil {
		return Entry{}, err
	}
	done, err := s.acquire(ctx, false, p)
	if err != nil {
		return Entry{}, err
	}
	defer done()
	f, proof, err := s.openProven(ctx, p, version)
	if err != nil {
		return Entry{}, err
	}
	defer f.Close()
	if err = s.checkProvenHandle(p, f, proof.info); err != nil {
		return Entry{}, err
	}
	out := entry(p, proof.info)
	out.Version = version
	return out, ctx.Err()
}

// ReadTransferChunk verifies complete covering blocks, even for unaligned
// ranges. Thus every returned byte belongs to the requested whole-file digest,
// without rehashing the whole file for every 64 KiB transport frame. The normal
// ReadChunk endpoint retains its stricter per-call whole-file revalidation.
func (s *Service) ReadTransferChunk(ctx context.Context, p, version string, offset int64, limit int) (Chunk, error) {
	out := Chunk{Path: p, Offset: offset}
	if err := s.validate(p, false); err != nil {
		return out, err
	}
	if offset < 0 || limit <= 0 || limit > s.opts.ChunkBytes {
		return out, ErrLimit
	}
	done, err := s.acquire(ctx, false, p)
	if err != nil {
		return out, err
	}
	defer done()
	f, proof, err := s.openProven(ctx, p, version)
	if err != nil {
		return out, err
	}
	defer f.Close()
	if offset > proof.info.Size() {
		return out, ErrTransfer
	}
	end := offset + min(int64(limit), proof.info.Size()-offset)
	out.Data = make([]byte, 0, int(end-offset))
	buffer := make([]byte, proofBlockBytes)
	for position := offset; position < end; {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		index := position / proofBlockBytes
		start := index * proofBlockBytes
		size := min(int64(proofBlockBytes), proof.info.Size()-start)
		block := buffer[:int(size)]
		if _, err = f.ReadAt(block, start); err != nil {
			return out, err
		}
		if sha256.Sum256(block) != proof.blocks[int(index)] {
			return out, ErrConflict
		}
		last := min(end, start+size)
		out.Data = append(out.Data, block[int(position-start):int(last-start)]...)
		position = last
	}
	if err = s.checkProvenHandle(p, f, proof.info); err != nil {
		return out, err
	}
	out.Version = version
	out.Total = proof.info.Size()
	out.Hash = hashBytes(out.Data)
	out.EOF = end == proof.info.Size()
	return out, ctx.Err()
}
