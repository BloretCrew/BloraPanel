package daemon

import (
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestContainerLogArchiveBoundedAndMarkedGap(t *testing.T) {
	d := &Daemon{config: Config{StateDir: t.TempDir()}}
	for i := 0; i < 105; i++ {
		w := model.ContainerLogWindow{ContainerID: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", ObservedAt: time.Unix(int64(i), 0), Frames: []model.ContainerLogFrame{{Stream: "stdout", Data: []byte("x")}}}
		if err := d.persistContainerLog(w); err != nil {
			t.Fatal(err)
		}
	}
	x, err := d.containerLogHistory("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(x) != 100 || !x[0].PossibleGap || x[0].ObservedAt.Unix() != 5 {
		t.Fatalf("archive=%d first=%+v", len(x), x[0])
	}
}

func TestContainerLogArchivePurgedAfterConfirmedDelete(t *testing.T) {
	d := &Daemon{config: Config{StateDir: t.TempDir()}}
	id := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	if err := d.persistContainerLog(model.ContainerLogWindow{ContainerID: id, Frames: []model.ContainerLogFrame{{Data: []byte("keep until delete")}}}); err != nil {
		t.Fatal(err)
	}
	if err := d.purgeContainerLogArchive(id); err != nil {
		t.Fatal(err)
	}
	if _, err := d.containerLogHistory(id, 1); err == nil {
		t.Fatal("archive remains after confirmed delete")
	}
}
