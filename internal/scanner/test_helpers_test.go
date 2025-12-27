package scanner

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"eclat/internal/feedback"
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

// MockNotifier - struktura pomocnicza tylko do testów
type MockNotifier struct {
	LastMsg   feedback.ToastField // Zapamiętujemy ostatnią wiadomość
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

// setupTestDB - Wersja PRO z Goose
func setupTestDB(t *testing.T) (*sql.DB, database.Querier) {
	// 1. Connection String (z fixem na czas)
	dsn := "file::memory:?cache=shared&_time_format=sqlite"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	// 2. GOOSE - Migracje
	// Ustawiamy dialekt na sqlite3 (Goose używa tego stringa nawet dla modernc)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("Failed to set goose dialect: %v", err)
	}

	// Wyłączamy logi Goose'a w testach, żeby nie śmiecił w terminalu
	goose.SetLogger(goose.NopLogger())

	// Odpalamy migracje z plików na dysku using relative path
	// Z katalogu internal/services musimy wyjść dwa razy w górę do sql/schema
	if err := goose.Up(db, "../../sql/schema"); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	queries := database.New(db)

	t.Cleanup(func() {
		db.Close()
	})

	return db, queries
}

// Helper do szybkiego wstawiania assetu do bazy (State Injection)
func insertTestAsset(t *testing.T, q database.Querier, folderID int64, path, hash string) database.Asset {
	ctx := context.Background()
	// Tworzymy asset z domyślnymi danymi
	params := database.CreateAssetParams{
		ScanFolderID:    sql.NullInt64{Int64: folderID, Valid: true},
		FilePath:        path,
		FileName:        filepath.Base(path),
		FileType:        "image",
		FileHash:        sql.NullString{String: hash, Valid: hash != ""},
		LastModified:    time.Now().Add(-1 * time.Hour), // Godzinę temu
		LastScanned:     time.Now().Add(-1 * time.Hour),
		HasAlphaChannel: sql.NullBool{Bool: false, Valid: true},
	}

	asset, err := q.CreateAsset(ctx, params)
	if err != nil {
		t.Fatalf("Failed to insert test asset: %v", err)
	}
	return asset
}

func setupLogicTest(t *testing.T) (*sql.DB, database.Querier, *Scanner, string) {
	conn, queries := setupTestDB(t) // Używamy Twojego helpera z setup_test.go
	root := t.TempDir()

	// Tworzymy folder skanowania w bazie
	_, err := queries.CreateScanFolder(context.Background(), root)
	assert.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockThumbGen := &MockThumbnailGenerator{}
	notifier := &MockNotifier{}
	scanner := NewScanner(conn, queries, mockThumbGen, logger, notifier)
	scanner.AddExtensions([]string{".txt", ".png"}) // Używamy prostych rozszerzeń
	return conn, queries, scanner, root
}

// --- Helpery bez zmian ---
func createDummyFile(t *testing.T, path string) {
	err := os.WriteFile(path, []byte("dummy content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file %s: %v", path, err)
	}
}

type MockThumbnailGenerator struct {
	ShouldFail bool
}

func (m *MockThumbnailGenerator) Generate(ctx context.Context, sourcePath string) (ThumbnailResult, error) {
	if m.ShouldFail {
		return ThumbnailResult{}, fmt.Errorf("mock error generator")
	}
	return ThumbnailResult{
		WebPath: "/thumbnails/mock_thumb.webp",
		Metadata: ImageMetadata{
			Width:           1920,
			Height:          1080,
			DominantColor:   "#FF0000",
			BitDepth:        8,
			HasAlphaChannel: false,
		},
		IsPlaceholder: false,
	}, nil
}
