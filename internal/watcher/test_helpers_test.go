package watcher

import (
	"context"
	"database/sql"
	"eclat/internal/config"
	"eclat/internal/database"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupTestDB initializes an in-memory SQLite database and applies migrations.
func setupTestDB(t *testing.T) (*sql.DB, database.Querier) {

	dsn := "file::memory:?cache=shared&_time_format=sqlite"

	db, err := sql.Open("sqlite", dsn)
	assert.NoError(t, err)

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}

	// Assuming the migration path is correct relative to the test location.
	if err := goose.Up(db, "../../sql/schema"); err != nil {
		t.Fatal("Failed to migrate DB:", err)
	}

	return db, database.New(db)
}

// setupWatcherTest prepares the main test environment for the watcher service.
// It returns the Service, Querier, RootDir (tmp), Context, and CancelFunc.
func setupWatcherTest(t *testing.T) (*Service, database.Querier, string, context.Context, context.CancelFunc) {
	// 1. Database
	_, queries := setupTestDB(t)

	// 2. Logger (discard output)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// 3. Temporary Directory
	rootDir := t.TempDir()

	// 4. Configuration (Create object and set with setter)
	cfg := config.NewScannerConfig()
	cfg.SetAllowedExtensions([]string{".png"}) // Set test extension

	// 5. Create Service (Inject config)
	svc, err := NewService(queries, logger, cfg)
	assert.NoError(t, err)

	// 6. Add root folder to DB so initFolders works
	ctx, cancel := context.WithCancel(context.Background())
	_, err = queries.CreateScanFolder(ctx, rootDir)
	assert.NoError(t, err)

	return svc, queries, rootDir, ctx, cancel
}

// createDummyFile helper creates a dummy file with content, ensuring the directory exists.
func createDummyFile(t *testing.T, path string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	err := os.WriteFile(path, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file %s: %v", path, err)
	}
}

// waitForEvent waits for a specific file path event on the channel within a timeout.
func waitForEvent(t *testing.T, ch <-chan string, expectedPath string, timeout time.Duration) {
	select {
	case received := <-ch:
		assert.Equal(t, expectedPath, received, "Received event for wrong path")
	case <-time.After(timeout):
		t.Fatalf("Timeout waiting for event: %s", expectedPath)
	}
}

// assertNoEvent ensures that no event is received on the channel within the duration.
func assertNoEvent(t *testing.T, ch <-chan string, duration time.Duration) {
	select {
	case event := <-ch:
		t.Fatalf("Unexpected event received: %s", event)
	case <-time.After(duration):
		// OK
	}
}
