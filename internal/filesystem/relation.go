package filesystem

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// RelationFacts is an internal Daemon/Master contract. It contains opaque
// fingerprints and actual filesystem object identities, never host paths.
// Ancestors includes the closest existing parent for a missing target.
type RelationFacts struct {
	MachineID       string   `json:"machineId"`
	RootID          string   `json:"rootId"`
	ObjectID        string   `json:"objectId,omitempty"`
	Ancestors       []string `json:"ancestors"`
	Exists          bool     `json:"exists"`
	PathFingerprint string   `json:"pathFingerprint"`
}

// OverlappingRelations also catches hardlinks and differently named roots.
// Object identities contain the volume/filesystem identity, so shared volumes
// can match even when reached by different Daemons or mount aliases.
func OverlappingRelations(a, b RelationFacts) bool {
	if a.ObjectID != "" && a.ObjectID == b.ObjectID {
		return true
	}
	for _, parent := range b.Ancestors {
		if a.ObjectID != "" && a.ObjectID == parent {
			return true
		}
	}
	for _, parent := range a.Ancestors {
		if b.ObjectID != "" && b.ObjectID == parent {
			return true
		}
	}
	return a.MachineID == b.MachineID && a.PathFingerprint == b.PathFingerprint
}

func (s *Service) RelationFacts(ctx context.Context, p string) (out RelationFacts, resultErr error) {
	if err := s.validate(p, true); err != nil {
		return out, err
	}
	done, err := s.acquire(ctx, false, p)
	if err != nil {
		return out, err
	}
	defer done()
	return InspectRootRelation(ctx, s.root, p)
}

// InspectRootRelation lets private backup repositories compare their own
// administrator-opened directory handle with an authorized instance service.
// It does not open a user supplied root or grant any filesystem access.
func InspectRootRelation(ctx context.Context, boundRoot *os.Root, p string) (out RelationFacts, resultErr error) {
	s := &Service{root: boundRoot}
	// Raw errors from administrative canonical-path checks may contain host
	// names. Only the safe capability-level error leaves this method.
	defer func() {
		if resultErr != nil && !errors.Is(resultErr, context.Canceled) && !errors.Is(resultErr, context.DeadlineExceeded) && !errors.Is(resultErr, ErrPath) && !errors.Is(resultErr, ErrConflict) {
			resultErr = ErrUnsupported
		}
	}()
	if err := s.validate(p, true); err != nil {
		return out, err
	}
	var err error
	if err = ctx.Err(); err != nil {
		return out, err
	}
	out.MachineID, err = relationMachineID()
	if err != nil {
		return out, err
	}
	root, err := s.root.Open(".")
	if err != nil {
		return out, err
	}
	defer root.Close()
	rootInfo, err := root.Stat()
	if err != nil {
		return out, err
	}
	out.RootID, err = relationObjectID(root)
	if err != nil {
		return out, err
	}
	canonical, err := filepath.EvalSymlinks(s.root.Name())
	if err != nil {
		return out, err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return out, err
	}
	currentRoot, err := os.Stat(canonical)
	if err != nil || !os.SameFile(rootInfo, currentRoot) {
		return out, ErrConflict
	}
	out.PathFingerprint = hashBytes([]byte(filepath.Join(canonical, filepath.FromSlash(p))))
	nearest := p
	var object *os.File
	for {
		object, err = s.root.OpenFile(nearest, os.O_RDONLY|readNonblock, 0)
		if err == nil {
			break
		}
		if !errors.Is(err, fs.ErrNotExist) || nearest == "." {
			return out, err
		}
		nearest = path.Dir(nearest)
	}
	defer object.Close()
	info, err := object.Stat()
	if err != nil {
		return out, err
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		return out, ErrUnsupported
	}
	if nearest != p && !info.IsDir() {
		return out, ErrConflict
	}
	actual, err := s.root.Lstat(nearest)
	if err != nil || !os.SameFile(info, actual) {
		return out, ErrConflict
	}
	id, err := relationObjectID(object)
	if err != nil {
		return out, err
	}
	full := filepath.Join(canonical, filepath.FromSlash(nearest))
	current, err := os.Stat(full)
	if err != nil || !os.SameFile(info, current) {
		return out, ErrConflict
	}
	if nearest == p {
		out.Exists, out.ObjectID = true, id
	} else {
		out.Ancestors = append(out.Ancestors, id)
	}
	ancestor := filepath.Dir(full)
	for ancestor != full {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		if len(out.Ancestors) >= 256 {
			return out, ErrLimit
		}
		f, err := os.Open(ancestor)
		if err != nil {
			return out, err
		}
		id, err := relationObjectID(f)
		closeErr := f.Close()
		if err != nil {
			return out, err
		}
		if closeErr != nil {
			return out, closeErr
		}
		out.Ancestors = append(out.Ancestors, id)
		full, ancestor = ancestor, filepath.Dir(ancestor)
	}
	// Detect namespace moves while collecting ancestor evidence. The rooted
	// handles remain the authority; stale canonical names fail closed.
	currentRoot, err = os.Stat(canonical)
	if err != nil || !os.SameFile(rootInfo, currentRoot) {
		return out, ErrConflict
	}
	current, err = os.Stat(filepath.Join(canonical, filepath.FromSlash(nearest)))
	if err != nil || !os.SameFile(info, current) {
		return out, ErrConflict
	}
	return out, nil
}

func (s *Service) expectedObject(p, id string) error {
	if id == "" {
		return nil
	}
	f, err := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	actual, err := relationObjectID(f)
	if err != nil {
		return err
	}
	if id != actual {
		return ErrConflict
	}
	return nil
}
