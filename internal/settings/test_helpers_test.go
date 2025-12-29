package settings

import (
	"context"
	"database/sql"
	"eclat/internal/config"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"eclat/internal/scanner"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite" // Driver
)

// MockNotifier captures notifications for testing assertions.
type MockNotifier struct {
	LastMsg   feedback.ToastField // Stores the last sent toast message
	CallCount int
	LastEvent string
}

func (m *MockNotifier) SendToast(ctx context.Context, msg feedback.ToastField) {
	m.LastMsg = msg
	m.CallCount++
}

func (m *MockNotifier) SendScannerStatus(ctx context.Context, status feedback.Status) {
	m.LastEvent = "scanner_status"
	m.CallCount++
}

func (m *MockNotifier) SendScanProgress(ctx context.Context, current, total int, message string) {
	m.LastEvent = "scan_progress"
	m.CallCount++
}
func (m *MockNotifier) EmitAssetsChanged(ctx context.Context) {
	m.LastEvent = "assets:changed"
	m.CallCount++
}

// setupTestDB initializes an in-memory SQLite database and applies migrations using Goose.
func setupTestDB(t *testing.T) (*sql.DB, database.Querier) {
	// 1. Connection String (with cache=shared for in-memory DBs)
	dsn := "file::memory:?cache=shared&_time_format=sqlite"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	// 2. GOOSE - Migrations
	// Force sqlite3 dialect (Goose uses this string even for modernc)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("Failed to set goose dialect: %v", err)
	}

	// Silence Goose logs during tests
	goose.SetLogger(goose.NopLogger())

	// Apply migrations from the disk using a relative path.
	// We need to navigate up from internal/settings to sql/schema.
	if err := goose.Up(db, "../../sql/schema"); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	queries := database.New(db)

	t.Cleanup(func() {
		db.Close()
	})

	return db, queries
}

// insertTestAsset helper quickly inserts a dummy asset into the database for testing state.
func insertTestAsset(t *testing.T, q database.Querier, folderID int64, path, hash string) database.Asset {
	ctx := context.Background()
	params := database.CreateAssetParams{
		ScanFolderID:    sql.NullInt64{Int64: folderID, Valid: true},
		FilePath:        path,
		FileName:        filepath.Base(path),
		FileType:        "image",
		FileHash:        sql.NullString{String: hash, Valid: hash != ""},
		LastModified:    time.Now().Add(-1 * time.Hour), // One hour ago
		LastScanned:     time.Now().Add(-1 * time.Hour),
		HasAlphaChannel: sql.NullBool{Bool: false, Valid: true},
	}

	asset, err := q.CreateAsset(ctx, params)
	if err != nil {
		t.Fatalf("Failed to insert test asset: %v", err)
	}
	return asset
}

// setupLogicTest prepares a complete environment for logic testing: DB, Queries, Scanner, and a Temp Directory.
func setupLogicTest(t *testing.T) (*sql.DB, database.Querier, *scanner.Scanner, string) {
	conn, queries := setupTestDB(t)
	root := t.TempDir()

	// Register the temp directory as a scan folder in the DB
	_, err := queries.CreateScanFolder(context.Background(), root)
	assert.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockThumbGen := &MockThumbnailGenerator{}
	notifier := &MockNotifier{}
	cfg := config.NewScannerConfig()
	cfg.SetAllowedExtensions([]string{".txt", ".png"})
	scanner := scanner.NewScanner(conn, queries, mockThumbGen, logger, notifier, cfg)
	return conn, queries, scanner, root
}

// createDummyFile helper creates a file with dummy content.
func createDummyFile(t *testing.T, path string) {
	err := os.WriteFile(path, []byte("dummy content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file %s: %v", path, err)
	}
}

// MockThumbnailGenerator mocks the thumbnail generation process.
type MockThumbnailGenerator struct {
	ShouldFail bool
}

func (m *MockThumbnailGenerator) Generate(ctx context.Context, sourcePath string) (scanner.ThumbnailResult, error) {
	if m.ShouldFail {
		return scanner.ThumbnailResult{}, fmt.Errorf("mock error generator")
	}
	return scanner.ThumbnailResult{
		WebPath: "/thumbnails/mock_thumb.webp",
		Metadata: scanner.ImageMetadata{
			Width:           1920,
			Height:          1080,
			DominantColor:   "#FF0000",
			BitDepth:        8,
			HasAlphaChannel: false,
		},
		IsPlaceholder: false,
	}, nil
}
