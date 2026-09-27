package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestMigrationRejectsFutureOrGappedHistoryWithoutChangingData(t *testing.T) {
	for _, change := range []string{"INSERT INTO schema_migrations(version) VALUES('999_future.sql')", "DELETE FROM schema_migrations WHERE version='002_node_settings.sql'"} {
		t.Run(change, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "state.db")
			s, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.DB.Exec("CREATE TABLE migration_canary(value TEXT); INSERT INTO migration_canary VALUES('retained')"); err != nil {
				t.Fatal(err)
			}
			if _, err = s.DB.Exec(change); err != nil {
				t.Fatal(err)
			}
			var count int
			if err = s.DB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&count); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			if opened, err := Open(path); !errors.Is(err, ErrSchemaVersion) {
				if opened != nil {
					opened.Close()
				}
				t.Fatalf("unsupported schema opened: %v", err)
			}
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var value string
			var after int
			if err = db.QueryRow("SELECT value FROM migration_canary").Scan(&value); err != nil || value != "retained" {
				t.Fatalf("canary changed: %q %v", value, err)
			}
			if err = db.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&after); err != nil || after != count {
				t.Fatalf("history changed: %d/%d %v", count, after, err)
			}
		})
	}
}

func TestMigrationAcceptsKnownPrefixAndAppliesRemainingVersions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := migrations.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("CREATE TABLE schema_migrations(version TEXT PRIMARY KEY);" + string(initial) + "; INSERT INTO schema_migrations VALUES('001_initial.sql')"); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var count int
	if err = s.DB.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 3 {
		t.Fatalf("known prefix did not migrate: %d %v", count, err)
	}
}
