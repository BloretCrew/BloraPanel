package filesystem

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"time"
)

// PruneRecords removes only terminal checkpoint metadata older than before.
// Active uploads are never expired by this method; their source can still be
// reselected. Callers retain the authoritative task result before pruning.
func (s *Service) PruneRecords(ctx context.Context, before time.Time) (Result, error) {
	r := Result{Stage: "pruning_metadata", Completed: []string{}}
	done, e := s.acquire(ctx, true, "@uploads", "@trash")
	if e != nil {
		return r, e
	}
	defer done()
	for _, dir := range []string{"uploads", "trash"} {
		f, e := s.state.Open(dir)
		if e != nil {
			return r, e
		}
		visited := 0
		for {
			if e = ctx.Err(); e != nil {
				f.Close()
				return r, e
			}
			names, err := f.Readdirnames(128)
			for _, name := range names {
				visited++
				if visited > s.opts.MaxEntries+1024 {
					f.Close()
					return r, ErrLimit
				}
				remove := false
				if strings.HasPrefix(name, ".checkpoint-") && validID(strings.TrimPrefix(name, ".checkpoint-")) {
					fi, e := s.state.Lstat(dir + "/" + name)
					if e != nil {
						f.Close()
						return r, e
					}
					remove = fi.Mode().IsRegular() && fi.ModTime().Before(before)
				} else if strings.HasSuffix(name, ".json") && validID(strings.TrimSuffix(name, ".json")) {
					if dir == "uploads" {
						u, e := s.loadUpload(strings.TrimSuffix(name, ".json"))
						if e != nil {
							f.Close()
							return r, e
						}
						remove = (u.Stage == "committed" || u.Stage == "cancelled") && u.Updated.Before(before)
					} else {
						var t TrashItem
						if e = s.readState(dir+"/"+name, &t); e != nil {
							f.Close()
							return r, e
						}
						remove = (t.Stage == "restored" || t.Stage == "purged") && t.DeletedAt.Before(before)
						// Prepared records with no associated recycle object never
						// deleted a file, or were already restored before a crash.
						if t.Stage == "prepared" && t.DeletedAt.Before(before) {
							p, e := trashPath(t)
							if e != nil {
								f.Close()
								return r, e
							}
							_, e = s.root.Lstat(p)
							remove = errors.Is(e, fs.ErrNotExist)
							if e != nil && !remove {
								f.Close()
								return r, e
							}
						}
					}
				}
				if remove {
					if e = ctx.Err(); e != nil {
						f.Close()
						return r, e
					}
					if e = s.state.Remove(dir + "/" + name); e != nil {
						f.Close()
						return r, e
					}
					r.Completed = append(r.Completed, dir+"/"+name)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				f.Close()
				return r, err
			}
		}
		f.Close()
		if e = syncDir(s.state, dir); e != nil {
			return r, e
		}
	}
	r.Stage = "committed"
	return r, nil
}
