package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"eclat/internal/feedback"
	"eclat/internal/scanner"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// setupTestDB initializes an in-memory SQLite database and applies migrations.
func setupTestDB(t *testing.T) (*sql.DB, database.Querier) {
	// Use a unique name for each test to avoid collisions in shared memory
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared&_time_format=sqlite"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("Failed to set goose dialect: %v", err)
	}
	goose.SetLogger(goose.NopLogger())

	// Apply migrations from sql/schema (relative to internal/app)
	if err := goose.Up(db, "../../sql/schema"); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	queries := database.New(db)
	t.Cleanup(func() {
		db.Close()
	})
	return db, queries
}

// setupAssetServiceTest creates an AssetService with a test DB and logger.
func setupAssetServiceTest(t *testing.T) (*AssetService, database.Querier) {
	sysDB, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	notifier := &MockNotifier{}
	service := NewAssetService(queries, sysDB, logger, notifier, "/tmp")
	service.Startup(context.Background())
	return service, queries
}

// setupMaterialSetServiceTest creates a MaterialSetService with a test DB and logger.
func setupMaterialSetServiceTest(t *testing.T) (*MaterialSetService, database.Querier) {
	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	thumbGen := &MockThumbGen{}
	service := NewMaterialSetService(queries, logger, thumbGen)
	service.Startup(context.Background())
	return service, queries
}

type MockThumbGen struct {
	ShouldFail bool
}

func (m *MockThumbGen) Generate(ctx context.Context, sourcePath string) (scanner.ThumbnailResult, error) {
	if m.ShouldFail {
		return scanner.ThumbnailResult{}, fmt.Errorf("mock error")
	}
	return scanner.ThumbnailResult{
		WebPath:       "/thumbnails/mock_cover.jpg",
		IsPlaceholder: false,
	}, nil
}

type MockNotifier struct {
	LastMsg   feedback.ToastField
	CallCount int
}

func (m *MockNotifier) SendToast(ctx context.Context, msg feedback.ToastField) {
	m.LastMsg = msg
	m.CallCount++
}

func (m *MockNotifier) SendScannerStatus(ctx context.Context, status feedback.Status) {
	m.CallCount++
}

func (m *MockNotifier) SendScanProgress(ctx context.Context, current, total int, lastFile string) {
	m.CallCount++
}

func (m *MockNotifier) EmitAssetsChanged(ctx context.Context) {
	m.CallCount++
}

// insertTestAsset creates a dummy asset in the database.
func insertTestAsset(t *testing.T, q database.Querier) database.Asset {
	return insertTestAssetWithParams(t, q, "test_file.png", "/tmp/test/test_file.png", false, false)
}

func insertTestAssetWithParams(t *testing.T, q database.Querier, fileName, filePath string, isDeleted, isHidden bool) database.Asset {
	return insertTestAssetWithParamsAndGroup(t, q, fileName, filePath, isDeleted, isHidden, "group_"+fileName)
}

func insertTestAssetWithParamsAndGroup(t *testing.T, q database.Querier, fileName, filePath string, isDeleted, isHidden bool, groupID string) database.Asset {
	ctx := context.Background()

	// Ensure a scan folder exists
	folder, err := q.CreateScanFolder(ctx, filepath.Dir(filePath))
	if err != nil {
		folder, err = q.GetScanFolderByPath(ctx, filepath.Dir(filePath))
		if err != nil {
			folder, _ = q.CreateScanFolder(ctx, "/tmp/random_"+fileName)
		}
	}

	params := database.CreateAssetParams{
		ScanFolderID:    sql.NullInt64{Int64: folder.ID, Valid: true},
		FileName:        fileName,
		FilePath:        filePath,
		FileType:        "image",
		FileSize:        1024,
		ThumbnailPath:   "",
		FileHash:        sql.NullString{String: "hash_" + fileName, Valid: true},
		ImageWidth:      sql.NullInt64{Int64: 100, Valid: true},
		ImageHeight:     sql.NullInt64{Int64: 100, Valid: true},
		DominantColor:   sql.NullString{String: "#000000", Valid: true},
		BitDepth:        sql.NullInt64{Int64: 8, Valid: true},
		HasAlphaChannel: sql.NullBool{Bool: false, Valid: true},
		LastModified:    time.Now(),
		LastScanned:     time.Now(),
		GroupID:         groupID,
	}

	asset, err := q.CreateAsset(ctx, params)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	if isDeleted {
		q.SoftDeleteAsset(ctx, asset.ID)
		asset.IsDeleted = true
	}
	if isHidden {
		q.SetAssetHidden(ctx, database.SetAssetHiddenParams{ID: asset.ID, IsHidden: true})
		asset.IsHidden = true
	}

	return asset
}
