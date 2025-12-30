package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
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
	dsn := "file::memory:?cache=shared&_time_format=sqlite"
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
	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewAssetService(queries, logger)
	service.Startup(context.Background())
	return service, queries
}

// insertTestAsset creates a dummy asset in the database.
func insertTestAsset(t *testing.T, q database.Querier) database.Asset {
	return insertTestAssetWithParams(t, q, "test_file.png", "/tmp/test/test_file.png", false, false)
}

func insertTestAssetWithParams(t *testing.T, q database.Querier, fileName, filePath string, isDeleted, isHidden bool) database.Asset {
	ctx := context.Background()

	// Ensure a scan folder exists
	folder, err := q.CreateScanFolder(ctx, filepath.Dir(filePath))
	if err != nil {
		// Try to fetch if it exists (e.g. from previous calls)
		folder, err = q.GetScanFolderByPath(ctx, filepath.Dir(filePath))
		if err != nil {
			// If still error, create a unique one
			folder, err = q.CreateScanFolder(ctx, filepath.Dir(filePath)+"_"+fileName)
			if err != nil {
				// Fallback to random
				folder, _ = q.CreateScanFolder(ctx, "/tmp/random_"+fileName)
			}
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
		GroupID:         "group_" + fileName,
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
