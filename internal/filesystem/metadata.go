package filesystem

import (
	"context"
	"os"
	"time"
)

// Metadata applies ordinary permission bits and times to an already-open
// object after its content/tree version check. Owner, ACL, special bits and
// symlinks are intentionally outside this capability. Partial errors remain
// errors; callers must not infer completion from a changed timestamp.
func (s *Service) Metadata(ctx context.Context, p, version string, mode uint32, modified time.Time) (Entry, error) {
	if mode > 0777 || modified.IsZero() {
		return Entry{}, ErrUnsupported
	}
	if err := s.validate(p, false); err != nil {
		return Entry{}, err
	}
	done, err := s.acquire(ctx, true, p)
	if err != nil {
		return Entry{}, err
	}
	defer done()
	if err = s.expected(ctx, p, version); err != nil {
		return Entry{}, err
	}
	f, err := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if err != nil {
		return Entry{}, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return Entry{}, err
	}
	if !before.IsDir() && !before.Mode().IsRegular() {
		return Entry{}, ErrUnsupported
	}
	current, err := s.root.Lstat(p)
	if err != nil {
		return Entry{}, err
	}
	if !os.SameFile(before, current) {
		return Entry{}, ErrConflict
	}
	if err = ctx.Err(); err != nil {
		return Entry{}, err
	}
	if err = applyHandleMetadata(f, mode, modified); err != nil {
		return Entry{}, err
	}
	after, err := s.root.Lstat(p)
	if err != nil {
		return Entry{}, err
	}
	if !os.SameFile(before, after) {
		return Entry{}, ErrConflict
	}
	return entry(p, after), nil
}
