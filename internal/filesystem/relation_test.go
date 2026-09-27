package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func relation(t *testing.T, s *Service, p string) RelationFacts {
	t.Helper()
	facts, err := s.RelationFacts(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	return facts
}

func TestRelationFactsDetectAliasesHardlinksAndNestedRoots(t *testing.T) {
	s, root, _ := testService(t, Options{})
	put(t, root, "source/file", "value")
	put(t, root, "sibling", "sibling")
	nested, err := New(filepath.Join(root, "source"), Options{StateDir: filepath.Join(t.TempDir(), "nested-private")})
	if err != nil {
		t.Fatal(err)
	}
	defer nested.Close()
	a := relation(t, s, "source/file")
	b := relation(t, nested, "file")
	if !OverlappingRelations(a, b) || a.ObjectID != b.ObjectID || a.MachineID != b.MachineID {
		t.Fatal("same actual file reached through distinct roots was not detected")
	}
	if !OverlappingRelations(relation(t, s, "source"), relation(t, nested, "new/destination")) {
		t.Fatal("nested missing destination not detected")
	}
	if !OverlappingRelations(relation(t, nested, "file"), relation(t, s, "source")) {
		t.Fatal("target ancestor not detected")
	}
	if OverlappingRelations(a, relation(t, s, "sibling")) {
		t.Fatal("independent sibling files rejected")
	}
	if err = os.Link(filepath.Join(root, "source/file"), filepath.Join(root, "hardlink")); err != nil {
		t.Fatal(err)
	}
	if !OverlappingRelations(a, relation(t, s, "hardlink")) {
		t.Fatal("hardlink alias not detected")
	}
	encoded, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), root) || strings.Contains(string(encoded), "source") {
		t.Fatal("host path leaked in relation facts")
	}
	if err = os.Symlink(filepath.Join(root, "source"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RelationFacts(context.Background(), "link/file"); err == nil {
		t.Fatal("relation facts followed a user symlink")
	}
}

func TestRelationFailsWhenCanonicalRootOrSourceObjectIsReplaced(t *testing.T) {
	s, root, _ := testService(t, Options{})
	put(t, root, "file", "same content")
	facts := relation(t, s, "file")
	v := version(t, s, "file")
	if err := os.Rename(filepath.Join(root, "file"), filepath.Join(root, "old-file")); err != nil {
		t.Fatal(err)
	}
	put(t, root, "file", "same content")
	if _, err := s.DeleteChecked(context.Background(), "file", v, facts.ObjectID); !errors.Is(err, ErrConflict) {
		t.Fatalf("same-content replacement deletion: %v", err)
	}
	content(t, root, "file", "same content")
	if err := os.Rename(root, root+"-renamed"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RelationFacts(context.Background(), "file"); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale root canonical name accepted: %v", err)
	}
}
