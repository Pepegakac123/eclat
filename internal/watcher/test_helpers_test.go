package watcher

import (
	"context"
	"database/sql"
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

func setupTestDB(t *testing.T) (*sql.DB, database.Querier) {

	dsn := "file::memory:?cache=shared&_time_format=sqlite"

	db, err := sql.Open("sqlite", dsn)
	assert.NoError(t, err)

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}

	if err := goose.Up(db, "../../sql/schema"); err != nil {
		t.Fatal("Failed to migrate DB:", err)
	}

	return db, database.New(db)
}

// setupWatcherTest - Główny setup testu
// Zwraca: Service, Querier, RootDir (tmp), Context, CancelFunc
func setupWatcherTest(t *testing.T) (*Service, database.Querier, string, context.Context, context.CancelFunc) {
	// 1. Baza danych
	_, queries := setupTestDB(t)

	// 2. Logger (wyciszony)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// 3. Folder tymczasowy
	rootDir := t.TempDir()

	// 4. Konfiguracja i utworzenie serwisu
	svc, err := NewService(queries, logger)
	assert.NoError(t, err)

	svc.config.AllowedExtensions = []string{".png"}

	// 5. Dodajemy folder root do bazy, żeby initFolders zadziałało
	ctx, cancel := context.WithCancel(context.Background())
	_, err = queries.CreateScanFolder(ctx, rootDir)
	assert.NoError(t, err)

	return svc, queries, rootDir, ctx, cancel
}

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

// waitForEvent - Helper czekający na zdarzenie z kanału (z timeoutem)
func waitForEvent(t *testing.T, ch <-chan string, expectedPath string, timeout time.Duration) {
	select {
	case received := <-ch:
		assert.Equal(t, expectedPath, received, "Otrzymano zdarzenie dla złej ścieżki")
	case <-time.After(timeout):
		t.Fatalf("Timeout: Nie otrzymano zdarzenia dla %s w czasie %v", expectedPath, timeout)
	}
}

// assertNoEvent - Helper upewniający się, że kanał milczy
func assertNoEvent(t *testing.T, ch <-chan string, waitTime time.Duration) {
	select {
	case received := <-ch:
		t.Fatalf("Otrzymano niespodziewane zdarzenie dla: %s", received)
	case <-time.After(waitTime):
		// Sukces - cisza
	}
}
