package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAssetService_GetAssetById(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	ctx := context.Background()

	// 1. Insert Asset
	asset := insertTestAsset(t, queries)

	// 2. Insert Tag
	tag, err := queries.CreateTag(ctx, "Nature")
	assert.NoError(t, err)
	err = queries.AddTagToAsset(ctx, database.AddTagToAssetParams{
		AssetID: asset.ID,
		TagID:   tag.ID,
	})
	assert.NoError(t, err)

	// 3. Test GetAssetById
	details, err := service.GetAssetById(asset.ID)
	assert.NoError(t, err)
	assert.Equal(t, asset.ID, details.ID)
	assert.Equal(t, asset.FileName, details.FileName)
	if assert.NotNil(t, details.GroupID) {
		assert.Equal(t, asset.GroupID, *details.GroupID)
	}
	assert.Contains(t, details.Tags, "Nature")
}

func TestAssetService_GetLibraryStats(t *testing.T) {
	service, queries := setupAssetServiceTest(t)

	// Insert 2 assets
	a1 := insertTestAssetWithParams(t, queries, "file1.png", "/tmp/1/file1.png", false, false)
	a2 := insertTestAssetWithParams(t, queries, "file2.png", "/tmp/1/file2.png", false, false)
	// Insert 1 deleted asset (should not count)
	insertTestAssetWithParams(t, queries, "deleted.png", "/tmp/1/deleted.png", true, false)

	// Verify they are visible
	assert.False(t, a1.IsHidden)
	assert.False(t, a2.IsHidden)

	stats, err := service.GetLibraryStats()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalAssets)
}

func TestAssetService_GetSidebarStats(t *testing.T) {
	service, queries := setupAssetServiceTest(t)

	// 1. Normal
	insertTestAssetWithParams(t, queries, "normal.png", "/tmp/normal.png", false, false)
	// 2. Favorite
	fav := insertTestAssetWithParams(t, queries, "fav.png", "/tmp/fav.png", false, false)
	service.ToggleAssetFavorite(fav.ID)
	// 3. Trash
	insertTestAssetWithParams(t, queries, "trash.png", "/tmp/trash.png", true, false)
	// 4. Hidden
	insertTestAssetWithParams(t, queries, "hidden.png", "/tmp/hidden.png", false, true)

	stats, err := service.GetSidebarStats()
	assert.NoError(t, err)

	assert.Equal(t, int64(3), stats.TotalAssets) // Normal + Favorite + Hidden? No, SQL usually excludes hidden from TotalAssets
	// Check SQL query in assets.sql:
	// AllCount: is_deleted=0 AND is_hidden=0
	// So: Normal (1) + Favorite (1) = 2. Hidden is excluded.

	// Wait, let's check my logic above:
	// normal.png: deleted=0, hidden=0 -> Counted
	// fav.png: deleted=0, hidden=0 -> Counted
	// trash.png: deleted=1 -> Not counted in AllCount
	// hidden.png: deleted=0, hidden=1 -> Not counted in AllCount (per query `AND is_hidden = 0`)

	assert.Equal(t, int64(3), stats.TotalAssets)
	assert.Equal(t, int64(1), stats.TotalFavorites)
	assert.Equal(t, int64(1), stats.TotalTrash)
	assert.Equal(t, int64(1), stats.TotalHidden)
}
func TestAssetService_SetAssetHidden(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	asset := insertTestAsset(t, queries)

	// Hide
	err := service.SetAssetHidden(asset.ID, true)
	assert.NoError(t, err)

	updated, _ := queries.GetAssetById(context.Background(), asset.ID)
	assert.True(t, updated.IsHidden)

	// Unhide
	err = service.SetAssetHidden(asset.ID, false)
	assert.NoError(t, err)

	updated, _ = queries.GetAssetById(context.Background(), asset.ID)
	assert.False(t, updated.IsHidden)
}

func TestAssetService_UpdateAssetMetadata(t *testing.T) {
	service, _ := setupAssetServiceTest(t)
	asset := insertTestAsset(t, service.db)

	desc := "New Description"
	rating := int64(4)
	fav := true

	req := UpdateAssetRequest{
		Description: &desc,
		Rating:      &rating,
		IsFavorite:  &fav,
	}

	updated, err := service.UpdateAssetMetadata(asset.ID, req)
	assert.NoError(t, err)
	assert.Equal(t, "New Description", updated.Description)
	assert.Equal(t, int64(4), updated.Rating)
	assert.True(t, updated.IsFavorite)

	// Invalid Rating
	badRating := int64(6)
	_, err = service.UpdateAssetMetadata(asset.ID, UpdateAssetRequest{Rating: &badRating})
	assert.Error(t, err)
}

func TestAssetService_SoftDeleteAndRestore(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	asset := insertTestAsset(t, queries)

	// Delete
	err := service.SoftDeleteAssets([]int64{asset.ID})
	assert.NoError(t, err)

	deleted, _ := queries.GetAssetById(context.Background(), asset.ID)
	assert.True(t, deleted.IsDeleted)

	// Restore
	err = service.RestoreAssets([]int64{asset.ID})
	assert.NoError(t, err)

	restored, _ := queries.GetAssetById(context.Background(), asset.ID)
	assert.False(t, restored.IsDeleted)
}

func TestAssetService_DeleteAssetsPermanently(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	asset := insertTestAsset(t, queries)

	err := service.DeleteAssetsPermanently([]int64{asset.ID})
	assert.NoError(t, err)

	_, err = queries.GetAssetById(context.Background(), asset.ID)
	assert.Error(t, err) // Should be not found
}

func TestAssetService_UpdateAssetType(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	asset := insertTestAsset(t, queries) // default type "image"

	// Valid switch
	err := service.UpdateAssetType(asset.ID, "texture")
	assert.NoError(t, err)

	updated, _ := queries.GetAssetById(context.Background(), asset.ID)
	assert.Equal(t, "texture", updated.FileType)

	// Invalid switch (to unsupported type)
	err = service.UpdateAssetType(asset.ID, "audio")
	assert.Error(t, err)
}

func TestAssetService_GetAssets(t *testing.T) {
	service, queries := setupAssetServiceTest(t)
	ctx := context.Background()

	// 1. Prepare Data
	// Asset 1: "Alpha", Rating 5, Favorite, Tag "Red"
	a1 := insertTestAssetWithParams(t, queries, "Alpha.png", "/tmp/Alpha.png", false, false)
	service.SetAssetRating(a1.ID, 5)
	service.ToggleAssetFavorite(a1.ID)

	tagRed, _ := queries.CreateTag(ctx, "Red")
	queries.AddTagToAsset(ctx, database.AddTagToAssetParams{AssetID: a1.ID, TagID: tagRed.ID})

	// Asset 2: "Beta", Rating 3, Not Favorite, Tag "Blue"
	a2 := insertTestAssetWithParams(t, queries, "Beta.png", "/tmp/Beta.png", false, false)
	service.SetAssetRating(a2.ID, 3)

	tagBlue, _ := queries.CreateTag(ctx, "Blue")
	queries.AddTagToAsset(ctx, database.AddTagToAssetParams{AssetID: a2.ID, TagID: tagBlue.ID})

	// Asset 3: "Gamma", Rating 1, Not Favorite, No Tags
	a3 := insertTestAssetWithParams(t, queries, "Gamma.png", "/tmp/Gamma.png", false, false)
	service.SetAssetRating(a3.ID, 1)

	// Asset 4: Deleted (Should not appear)
	insertTestAssetWithParams(t, queries, "Deleted.png", "/tmp/Deleted.png", true, false)

	// 2. Test Cases

	// Case A: No Filters (Should get 3 assets)
	res, err := service.GetAssets(AssetQueryFilters{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, res.TotalCount)
	assert.Len(t, res.Items, 3)

	// Case B: Search Query "Alpha"
	res, err = service.GetAssets(AssetQueryFilters{
		Page:     1,
		PageSize: 10,
		Query:    "Alpha",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, a1.FileName, res.Items[0].FileName)

	// Case C: Rating Range [4, 5] (Should get Alpha)
	res, err = service.GetAssets(AssetQueryFilters{
		Page:        1,
		PageSize:    10,
		RatingRange: []int{4, 5},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, a1.ID, res.Items[0].ID)

	// Case D: Only Favorites (Should get Alpha)
	res, err = service.GetAssets(AssetQueryFilters{
		Page:          1,
		PageSize:      10,
		OnlyFavorites: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, a1.ID, res.Items[0].ID)

	// Case E: Tag "Blue" (Should get Beta)
	res, err = service.GetAssets(AssetQueryFilters{
		Page:     1,
		PageSize: 10,
		Tags:     []string{"Blue"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, a2.ID, res.Items[0].ID)
}

func TestAssetService_GetAssets_ByFileType(t *testing.T) {
	service, queries := setupAssetServiceTest(t)

	// 1. Prepare Data
	insertTestAssetWithParams(t, queries, "image.png", "/tmp/image.png", false, false) // default "image"
	
	// Insert a Model
	folder, _ := queries.CreateScanFolder(context.Background(), "/tmp/models")
	queries.CreateAsset(context.Background(), database.CreateAssetParams{
		ScanFolderID: sql.NullInt64{Int64: folder.ID, Valid: true},
		FileName:     "model.blend",
		FilePath:     "/tmp/models/model.blend",
		FileType:     "model",
		FileSize:     2048,
		LastModified: time.Now(),
		LastScanned:  time.Now(),
		GroupID:      "group_model",
	})

	// Insert a Texture
	queries.CreateAsset(context.Background(), database.CreateAssetParams{
		ScanFolderID: sql.NullInt64{Int64: folder.ID, Valid: true},
		FileName:     "texture.tga",
		FilePath:     "/tmp/models/texture.tga",
		FileType:     "texture",
		FileSize:     512,
		LastModified: time.Now(),
		LastScanned:  time.Now(),
		GroupID:      "group_texture",
	})

	// 2. Test Cases

	// Case A: All (should get 3)
	res, err := service.GetAssets(AssetQueryFilters{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 3, res.TotalCount)

	// Case B: Filter by "model"
	res, err = service.GetAssets(AssetQueryFilters{
		Page:      1,
		PageSize:  10,
		FileTypes: []string{"model"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, "model", res.Items[0].FileType)
	assert.Equal(t, "model.blend", res.Items[0].FileName)

	// Case C: Filter by "texture"
	res, err = service.GetAssets(AssetQueryFilters{
		Page:      1,
		PageSize:  10,
		FileTypes: []string{"texture"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, "texture", res.Items[0].FileType)

	// Case D: Filter by "model" AND "texture"
	res, err = service.GetAssets(AssetQueryFilters{
		Page:      1,
		PageSize:  10,
		FileTypes: []string{"model", "texture"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, res.TotalCount)
}
