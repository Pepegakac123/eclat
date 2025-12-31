package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaterialSetService_CreateAndGet(t *testing.T) {
	service, _ := setupMaterialSetServiceTest(t)

	name := "Sci-Fi Collection"
	desc := "A collection of sci-fi assets"
	color := "#FF0000"

	req := CreateMaterialSetRequest{
		Name:        name,
		Description: &desc,
		CustomColor: &color,
	}

	// 1. Create
	ms, err := service.Create(req)
	assert.NoError(t, err)
	assert.NotZero(t, ms.ID)
	assert.Equal(t, name, ms.Name)
	assert.Equal(t, desc, *ms.Description)
	assert.Equal(t, color, *ms.CustomColor)

	// 2. GetById
	fetched, err := service.GetById(ms.ID)
	assert.NoError(t, err)
	assert.Equal(t, ms.ID, fetched.ID)
	assert.Equal(t, name, fetched.Name)
}

func TestMaterialSetService_GetAll(t *testing.T) {
	service, _ := setupMaterialSetServiceTest(t)

	// Create 2 sets
	_, _ = service.Create(CreateMaterialSetRequest{Name: "Set 1"})
	_, _ = service.Create(CreateMaterialSetRequest{Name: "Set 2"})

	all, err := service.GetAll()
	assert.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestMaterialSetService_Update(t *testing.T) {
	service, _ := setupMaterialSetServiceTest(t)

	ms, _ := service.Create(CreateMaterialSetRequest{Name: "Original Name"})

	newName := "Updated Name"
	req := CreateMaterialSetRequest{
		Name: newName,
	}

	updated, err := service.Update(ms.ID, req)
	assert.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	fetched, _ := service.GetById(ms.ID)
	assert.Equal(t, newName, fetched.Name)
}

func TestMaterialSetService_Delete(t *testing.T) {
	service, _ := setupMaterialSetServiceTest(t)

	ms, _ := service.Create(CreateMaterialSetRequest{Name: "To Delete"})

	err := service.Delete(ms.ID)
	assert.NoError(t, err)

	_, err = service.GetById(ms.ID)
	assert.Error(t, err)
}

func TestMaterialSetService_AddAndRemoveAsset(t *testing.T) {
	msService, queries := setupMaterialSetServiceTest(t)
	// We need AssetService helpers to create assets easily
	asset := insertTestAsset(t, queries)

	ms, _ := msService.Create(CreateMaterialSetRequest{Name: "Test Collection"})

	// 1. Add Asset
	err := msService.AddAsset(ms.ID, asset.ID)
	assert.NoError(t, err)

	// Verify total assets count
	updated, _ := msService.GetById(ms.ID)
	assert.Equal(t, int64(1), updated.TotalAssets)

	// 2. Remove Asset
	err = msService.RemoveAsset(ms.ID, asset.ID)
	assert.NoError(t, err)

	updated, _ = msService.GetById(ms.ID)
	assert.Equal(t, int64(0), updated.TotalAssets)
}

func TestMaterialSetService_SetMaterialSetCoverFromFile(t *testing.T) {
	service, _ := setupMaterialSetServiceTest(t)

	ms, _ := service.Create(CreateMaterialSetRequest{Name: "Collection with Cover"})

	// Set cover from file (mocked thumb gen will return /thumbnails/mock_cover.jpg)
	updated, err := service.SetMaterialSetCoverFromFile(ms.ID, "/path/to/image.jpg")
	assert.NoError(t, err)
	assert.NotNil(t, updated.CustomCoverUrl)
	assert.Equal(t, "/thumbnails/mock_cover.jpg", *updated.CustomCoverUrl)
	assert.Equal(t, "/thumbnails/mock_cover.jpg", updated.ThumbnailPath)
}
