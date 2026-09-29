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
	script := `#!/bin/sh
case "$1" in
list-units) printf '%s\n' 'z.service loaded active running Last' 'm.service loaded active running Middle';;
list-unit-files) printf '%s\n' 'm.service enabled enabled' 'a.service disabled enabled' 'unused@.service disabled enabled';;
show)
  shift
  while [ "$1" != '--' ]; do shift; done
  shift
  for unit in "$@"; do
    case "$unit" in
      a.service) state=inactive; description=First;;
      m.service) state=active; description=Middle;;
      z.service) state=active; description=Last;;
      *) exit 23;;
    esac
    printf 'Id=%s\nNames=%s\nActiveState=%s\nLoadState=loaded\nDescription=%s\n\n' "$unit" "$unit" "$state" "$description"
  done;;
*) exit 24;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "systemctl"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	first, err := ListServicePage(context.Background(), 2, "")
	if err != nil || len(first.Items) != 2 || first.Items[0].Name != "a.service" || first.NextAfter != "m.service" {
		t.Fatalf("bad first page %+v %v", first, err)
	}
	if first.Items[0].State != "inactive" || first.Items[0].Description != "First" {
		t.Fatalf("unloaded installed service lost its actual properties: %+v", first.Items[0])
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
	script.WriteString("#!/bin/sh\ncase \"$1\" in\nlist-units|list-unit-files) printf '%s\\n'")
	for i := 204; i >= 0; i-- {
		fmt.Fprintf(&script, " 'task%03d.timer disabled enabled'", i)
	}
	script.WriteString(`
;;
list-timers) exit 0;;
show)
 shift
 while [ "$1" != '--' ]; do shift; done
 shift
 for unit in "$@"; do
  printf 'Id=%s\nNames=%s\nLoadState=loaded\nUnitFileState=disabled\nTimersMonotonic={ OnActiveUSec=1h }\n\n' "$unit" "$unit"
 done;;
*) exit 24;;
esac
`)
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
			if item.State != "disabled" || item.Schedule == "" {
				t.Fatalf("unloaded timer lost enablement or schedule: %+v", item)
			}
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
