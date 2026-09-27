//go:build linux

package systeminfo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServicePagesReachLaterNames(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	script := "#!/bin/sh\nprintf '%s\\n' 'z.service loaded active running Last' 'a.service loaded inactive dead First' 'm.service loaded active running Middle'\n"
	if err := os.WriteFile(filepath.Join(dir, "systemctl"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	first, err := ListServicePage(context.Background(), 2, "")
	if err != nil || len(first.Items) != 2 || first.Items[0].Name != "a.service" || first.NextAfter != "m.service" {
		t.Fatalf("bad first page %+v %v", first, err)
	}
	last, err := ListServicePage(context.Background(), 2, first.NextAfter)
	if err != nil || len(last.Items) != 1 || last.Items[0].Name != "z.service" || last.NextAfter != "" {
		t.Fatalf("bad last page %+v %v", last, err)
	}
	if _, err := ListServicePage(context.Background(), 2, "../bad"); err == nil {
		t.Fatal("invalid cursor accepted")
	}
}

func TestScheduledTaskPagesBeyondTwoHundred(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	var script strings.Builder
	script.WriteString("#!/bin/sh\nprintf '%s\\n'")
	for i := 204; i >= 0; i-- {
		fmt.Fprintf(&script, " 'n/a n/a n/a n/a task%03d.timer task%03d.service'", i, i)
	}
	script.WriteString("\n")
	if err := os.WriteFile(filepath.Join(dir, "systemctl"), []byte(script.String()), 0700); err != nil {
		t.Fatal(err)
	}
	after := ""
	seen := map[string]bool{}
	for pages := 0; pages < 3; pages++ {
		page, err := ListScheduledTaskPage(context.Background(), 100, after)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range page.Items {
			if seen[item.Name] || item.Name <= after {
				t.Fatalf("duplicate or shifted identity %s", item.Name)
			}
			seen[item.Name] = true
		}
		after = page.NextAfter
		if pages < 2 && after == "" {
			t.Fatal("truncated list")
		}
	}
	if len(seen) != 205 || after != "" {
		t.Fatalf("missing tasks %d, cursor %s", len(seen), after)
	}
}
