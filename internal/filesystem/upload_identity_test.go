package filesystem

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestUploadCommitChecksTargetAndCommittedObjectIdentity(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	if _, err := s.WriteText(ctx, "destination", "baseline", MissingVersion); err != nil {
		t.Fatal(err)
	}
	relation, err := s.RelationFacts(ctx, "destination")
	if err != nil {
		t.Fatal(err)
	}
	spec := uploadSpec([]byte("new data"), "destination")
	spec.ExpectedVersion = hashBytes([]byte("baseline"))
	spec.ExpectedObjectID = relation.ObjectID
	sendAll(t, s, spec, []byte("new data"))
	if err = os.WriteFile(root+"/replacement", []byte("baseline"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(root+"/replacement", root+"/destination"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CommitUpload(ctx, spec.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("same content object replacement %v", err)
	}
	spec = uploadSpec([]byte("same"), "same")
	u := sendAll(t, s, spec, []byte("same"))
	part, err := s.root.Open(uploadPart(u))
	if err != nil {
		t.Fatal(err)
	}
	u.CommitObjectID, err = relationObjectID(part)
	part.Close()
	if err != nil {
		t.Fatal(err)
	}
	u.Stage = "committing"
	if err = s.writeState(uploadRecord(u.Spec.ID), u); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(root+"/same", []byte("same"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.root.Remove(uploadPart(u)); err != nil {
		t.Fatal(err)
	}
	if status, err := s.UploadStatus(ctx, spec.ID); err == nil || status.Stage == "committed" {
		t.Fatalf("unrelated same hash certified %+v %v", status, err)
	}
}
