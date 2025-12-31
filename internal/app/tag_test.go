package app

import (
	"context"
	"eclat/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagService_GetAll(t *testing.T) {
	_, queries := setupTestDB(t)
	ctx := context.Background()

	// 1. Create tags
	tag1, _ := queries.CreateTag(ctx, "Nature")
	tag2, _ := queries.CreateTag(ctx, "City")

	// 2. Create assets and link tags
	a1 := insertTestAssetWithParams(t, queries, "file1.png", "/tmp/test/file1.png", false, false)
	_ = queries.AddTagToAsset(ctx, database.AddTagToAssetParams{AssetID: a1.ID, TagID: tag1.ID})
	_ = queries.AddTagToAsset(ctx, database.AddTagToAssetParams{AssetID: a1.ID, TagID: tag2.ID})

	a2 := insertTestAssetWithParams(t, queries, "file2.png", "/tmp/test/file2.png", false, false)
	_ = queries.AddTagToAsset(ctx, database.AddTagToAssetParams{AssetID: a2.ID, TagID: tag1.ID})

	// 3. Test TagService
	service := NewTagService(queries, nil)
	service.Startup(ctx)

	tags, err := service.GetAll()
	assert.NoError(t, err)
	assert.Len(t, tags, 2)

	// Verify counts
	for _, tg := range tags {
		if tg.Name == "Nature" {
			assert.Equal(t, int64(2), tg.AssetCount)
		} else if tg.Name == "City" {
			assert.Equal(t, int64(1), tg.AssetCount)
		}
	}
}
