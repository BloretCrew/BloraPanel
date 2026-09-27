package storage

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const metadataCrashFixtureEnv = "BLORA_METADATA_TX_CRASH_DB"
const metadataCrashReadyLine = "METADATA_TX_WRITES_READY"

// TestMetadataMutationRollsBackOnProcessKillBeforeCommit kills a separate
// process after the mutation callback has written related records but before
// metadataMutation writes its idempotency receipt, audit row, and commits.
func TestMetadataMutationRollsBackOnProcessKillBeforeCommit(t *testing.T) {
	if path := os.Getenv(metadataCrashFixtureEnv); path != "" {
		runMetadataMutationCrashHelper(t, path)
		return
	}

	dbPath := filepath.Join(t.TempDir(), "state.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.DB.ExecContext(context.Background(), `INSERT INTO users(id,name,password_hash,admin,disabled) VALUES(?,?,?,?,?)`, "crash-admin", "crash-admin", []byte("test-only"), true, false); err != nil {
		_ = store.Close()
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMetadataMutationRollsBackOnProcessKillBeforeCommit$", "-test.count=1")
	cmd.Env = append(os.Environ(), metadataCrashFixtureEnv+"="+dbPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	ready := make(chan struct{}, 1)
	scanDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		announced := false
		for scanner.Scan() {
			if scanner.Text() == metadataCrashReadyLine && !announced {
				announced = true
				ready <- struct{}{}
			}
		}
		scanDone <- scanner.Err()
	}()
	select {
	case <-ready:
	case scanErr := <-scanDone:
		waitErr := cmd.Wait()
		t.Fatalf("mutation helper exited before uncommitted writes: scan=%v wait=%v stderr=%s", scanErr, waitErr, stderr.String())
	case <-time.After(10 * time.Second):
		t.Fatalf("mutation helper did not reach the pre-commit crash point; stderr=%s", stderr.String())
	}

	if err = cmd.Process.Kill(); err != nil {
		t.Fatalf("kill mutation helper: %v", err)
	}
	if err = cmd.Wait(); err == nil {
		t.Fatal("mutation helper exited normally instead of being killed")
	} else {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("wait for killed mutation helper: %v", err)
		}
	}
	if scanErr := <-scanDone; scanErr != nil {
		t.Fatalf("read helper output after process kill: %v", scanErr)
	}

	store, err = Open(dbPath)
	if err != nil {
		t.Fatalf("reopen SQLite after process kill: %v", err)
	}
	defer store.Close()
	ctx := context.Background()
	for _, record := range [][2]string{{"platform_schedules_v1", "schedule-crash-fixture"}, {"platform_schedule_fires_v1", "fire-crash-fixture"}, {"metadata_request", "crash-admin:crash-request"}} {
		var revision int64
		err = store.DB.QueryRowContext(ctx, "SELECT revision FROM records WHERE namespace=? AND id=?", record[0], record[1]).Scan(&revision)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("interrupted transaction left %s/%s revision=%d err=%v", record[0], record[1], revision, err)
		}
	}
	var auditCount int
	if err = store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit WHERE request_id=?", "crash-request").Scan(&auditCount); err != nil || auditCount != 0 {
		t.Fatalf("interrupted transaction left audit rows: count=%d err=%v", auditCount, err)
	}
	var adminCount int
	if err = store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id=? AND admin=1", "crash-admin").Scan(&adminCount); err != nil || adminCount != 1 {
		t.Fatalf("committed state did not survive recovery: count=%d err=%v", adminCount, err)
	}
	var integrity string
	if err = store.DB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("SQLite integrity after process kill: result=%q err=%v", integrity, err)
	}
}

func runMetadataMutationCrashHelper(t *testing.T, dbPath string) {
	store, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.MetadataMutation(context.Background(), "crash-admin", "crash-request", "schedule.save", map[string]string{"id": "schedule-crash-fixture"}, func(tx *sql.Tx) (any, error) {
		if _, err := tx.ExecContext(context.Background(), "INSERT INTO records(namespace,id,revision,document) VALUES(?,?,1,?)", "platform_schedules_v1", "schedule-crash-fixture", []byte(`{"pending":"fixture"}`)); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(context.Background(), "INSERT INTO records(namespace,id,revision,document) VALUES(?,?,1,?)", "platform_schedule_fires_v1", "fire-crash-fixture", []byte(`{"state":"accepted"}`)); err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(os.Stdout, metadataCrashReadyLine); err != nil {
			return nil, err
		}
		select {}
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("crash callback returned and transaction committed")
}
