package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type AssetService struct {
	ctx    context.Context
	db     database.Querier
	logger *slog.Logger
}

func NewAssetService(db database.Querier, logger *slog.Logger) *AssetService {
	return &AssetService{
		db:     db,
		logger: logger,
	}
}

func (s *AssetService) Startup(ctx context.Context) {
	s.ctx = ctx
}

// ==========================================
// Data Transfer Objects (DTOs) for Wails
// ==========================================

// AssetDetails to pełny obiekt assetu zwracany do UI.
// Odzwierciedla to, co frontend potrzebuje w Inspektorze.
type AssetDetails struct {
	ID            int64     `json:"id"`
	FilePath      string    `json:"filePath"`
	FileName      string    `json:"fileName"`
	FileType      string    `json:"fileType"`
	ThumbnailPath string    `json:"thumbnailPath"`
	DateAdded     time.Time `json:"dateAdded"`
	LastModified  time.Time `json:"lastModified"`
	FileSize      int64     `json:"fileSize"`

	// Metadane edytowalne
	Rating      int64  `json:"rating"`
	IsFavorite  bool   `json:"isFavorite"`
	Description string `json:"description"`
	IsDeleted   bool   `json:"isDeleted"`
	IsHidden    bool   `json:"isHidden"`

	// Relacje i Grupowanie
	GroupID       string   `json:"groupId"`
	Tags          []string `json:"tags"`
	MaterialSets  []string `json:"materialSets"`
	DominantColor string   `json:"dominantColor"`
}

type AssetSibling struct {
	ID       int64  `json:"id"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
}

// AssetQueryFilters odzwierciedla obiekt filtrów z useGalleryStore.ts
type AssetQueryFilters struct {
	// Paginacja
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`

	// Wyszukiwanie i Tagi
	Query        string   `json:"searchQuery"`
	Tags         []string `json:"tags"`
	MatchAllTags bool     `json:"matchAllTags"`

	// Listy
	FileTypes []string `json:"fileTypes"`
	Colors    []string `json:"colors"`

	// Zakresy (Ranges)
	RatingRange   []int `json:"ratingRange"`   // [min, max]
	WidthRange    []int `json:"widthRange"`    // [min, max]
	HeightRange   []int `json:"heightRange"`   // [min, max]
	FileSizeRange []int `json:"fileSizeRange"` // [min, max] w MB

	// Data
	DateRange struct {
		From *string `json:"from"`
		To   *string `json:"to"`
	} `json:"dateRange"`

	// Specjalne
	HasAlpha      *bool `json:"hasAlpha"` // null = wszystkie, true = z alpha, false = bez
	OnlyFavorites bool  `json:"onlyFavorites"`

	// Kontekst
	IsDeleted    bool   `json:"isDeleted"`
	IsHidden     bool   `json:"isHidden"`
	CollectionID *int64 `json:"collectionId"`

	// Sortowanie
	SortOption string `json:"sortOption"` // "dateadded", "filename", "filesize", "lastmodified", "rating"
	SortDesc   bool   `json:"sortDesc"`
}

// PagedAssetResult to wynik paginacji dla wirtualnej listy.
type PagedAssetResult struct {
	Items      []AssetDetails `json:"items"`
	TotalCount int            `json:"totalCount"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
}

// LibraryStats to statystyki globalne biblioteki.
type LibraryStats struct {
	TotalAssets int64      `json:"totalAssets"`
	TotalSize   int64      `json:"totalSize"`
	LastScan    *time.Time `json:"lastScan"` // Pointer, bo może być null
}

// SidebarStats to liczniki dla paska bocznego.
type SidebarStats struct {
	TotalAssets        int64 `json:"totalAssets"`
	TotalUncategorized int64 `json:"totalUncategorized"`
	TotalFavorites     int64 `json:"totalFavorites"`
	TotalTrash         int64 `json:"totalTrash"`
	TotalHidden        int64 `json:"totalHidden"` // Nowe pole
}

type UpdateAssetRequest struct {
	Description *string
	Rating      *int64
	IsFavorite  *bool
}

// ==========================================
// Asset Logic Implementation
// ==========================================

// GetAssetById pobiera pojedynczy asset wraz z jego tagami i groupID.
func (s *AssetService) GetAssetById(id int64) (*AssetDetails, error) {
	ctx := context.Background()

	// 1. Pobierz asset (sqlc wygeneruje metodę zwracającą struct z nowymi polami)
	asset, err := s.db.GetAssetById(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Pobierz tagi
	tagNames, err := s.db.GetTagsNamesByAssetID(ctx, asset.ID)
	if err != nil {
		return nil, err
	}

	// 3. Mapowanie
	details := &AssetDetails{
		ID:            asset.ID,
		FilePath:      asset.FilePath,
		FileName:      asset.FileName,
		FileType:      asset.FileType,
		ThumbnailPath: asset.ThumbnailPath,
		DateAdded:     asset.DateAdded,
		LastModified:  asset.LastModified,
		FileSize:      asset.FileSize,
		Rating:        asset.Rating,
		IsFavorite:    asset.IsFavorite.Bool,
		Description:   asset.Description.String,
		IsDeleted:     asset.IsDeleted,
		IsHidden:      asset.IsHidden,
		GroupID:       asset.GroupID,
		Tags:          tagNames,
		DominantColor: asset.DominantColor.String,
	}

	return details, nil
}

// GetLibraryStats zwraca ogólne statystyki (liczba plików, rozmiar).
func (s *AssetService) GetLibraryStats() (*LibraryStats, error) {
	ctx := context.Background()
	stats, err := s.db.GetLibraryStats(ctx)
	if err != nil {
		return nil, err
	}

	// Konwersja z sql.NullTime (interface{})
	var lastScan *time.Time
	if stats.LastScan != nil {
		if t, ok := stats.LastScan.(time.Time); ok {
			lastScan = &t
		}
	}

	return &LibraryStats{
		TotalAssets: stats.TotalCount,
		TotalSize:   stats.TotalSize,
		LastScan:    lastScan,
	}, nil
}

// GetSidebarStats zwraca liczniki dla menu bocznego (All, Favorites, Trash, Hidden, Uncategorized).
func (s *AssetService) GetSidebarStats() (*SidebarStats, error) {
	ctx := context.Background()

	// Query teraz zwraca jeden wiersz z 5 kolumnami
	stats, err := s.db.GetSidebarStats(ctx)
	if err != nil {
		return nil, err
	}

	return &SidebarStats{
		TotalAssets:        stats.AllCount,
		TotalFavorites:     stats.FavoritesCount,
		TotalTrash:         stats.TrashCount,
		TotalHidden:        stats.HiddenCount,
		TotalUncategorized: stats.UncategorizedCount,
	}, nil
}

// SetAssetHidden ustawia flagę ukrycia dla assetu.
func (s *AssetService) SetAssetHidden(id int64, hidden bool) error {
	ctx := context.Background()
	return s.db.SetAssetHidden(ctx, database.SetAssetHiddenParams{
		IsHidden: hidden,
		ID:       id,
	})
}

func (s *AssetService) UpdateAssetMetadata(id int64, req UpdateAssetRequest) (*AssetDetails, error) {

	if req.Rating != nil {
		if *req.Rating < 0 || *req.Rating > 5 {
			return nil, errors.New("rating must be between 0 and 5")
		}
	}

	if req.Description != nil {
		if len(*req.Description) > 500 {
			return nil, errors.New("description too long (max 500 chars)")
		}
	}

	params := database.UpdateAssetMetadataParams{
		ID: id,
		Description: sql.NullString{
			String: getStringValue(req.Description),
			Valid:  req.Description != nil,
		},
		Rating: sql.NullInt64{
			Int64: getInt64Value(req.Rating),
			Valid: req.Rating != nil,
		},
		IsFavorite: sql.NullBool{
			Bool:  getBoolValue(req.IsFavorite),
			Valid: req.IsFavorite != nil,
		},
		ThumbnailPath: sql.NullString{Valid: false},
	}

	updatedAsset, err := s.db.UpdateAssetMetadata(s.ctx, params)
	if err != nil {
		return nil, err
	}

	return s.GetAssetById(updatedAsset.ID)
}

// Helpery, żeby nie pisać if-ów w kółko
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func getInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// ToggleAssetFavorite przełącza flagę ulubionego
func (s *AssetService) ToggleAssetFavorite(id int64) error {
	return s.db.ToggleAssetFavorite(s.ctx, id)
}

// SetAssetRating ustawia ocenę.
func (s *AssetService) SetAssetRating(id int64, rating int64) error {
	if rating < 0 || rating > 5 {
		return errors.New("rating must be between 0 and 5")
	}
	return s.db.SetAssetRating(s.ctx, database.SetAssetRatingParams{
		Rating: rating,
		ID:     id,
	})
}

// SoftDeleteAssets przenosi assety do kosza.
func (s *AssetService) SoftDeleteAssets(ids []int64) error {
	const batchSize = 500
	for i := 0; i < len(ids); i += batchSize {
		end := min(i+batchSize, len(ids))
		err := s.db.SoftDeleteAssets(s.ctx, ids[i:end])
		if err != nil {
			return err
		}
	}
	return nil
}

// RestoreAssets przywraca assety z kosza.
func (s *AssetService) RestoreAssets(ids []int64) error {
	const batchSize = 500
	for i := 0; i < len(ids); i += batchSize {
		end := min(i+batchSize, len(ids))
		err := s.db.RestoreAssets(s.ctx, ids[i:end])
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAssetsPermanently usuwa assety z bazy na zawsze.
func (s *AssetService) DeleteAssetsPermanently(ids []int64) error {
	// TODO: Tutaj powinieneś też usunąć plik z dysku fizycznego, jeśli taka jest wola usera!
	// Na razie tylko baza.
	for _, id := range ids {
		if err := s.db.DeleteAssetPermanent(s.ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// GetAssetVersions zwraca wszystkie assety należące do tej samej grupy co podany ID.
func (s *AssetService) GetAssetVersions(assetId int64) ([]AssetDetails, error) {
	asset, err := s.db.GetAssetById(s.ctx, assetId)
	if err != nil {
		return nil, err
	}

	if asset.GroupID == "" {
		return []AssetDetails{}, nil
	}
	siblings, err := s.db.GetAssetsByGroupID(s.ctx, asset.GroupID)
	if err != nil {
		return nil, err
	}

	// 3. Mapowanie na AssetDetails (skrócone, bez tagów dla wydajności listy)
	var results []AssetDetails
	for _, sib := range siblings {
		results = append(results, AssetDetails{
			ID:       sib.ID,
			FileName: sib.FileName,
			FilePath: sib.FilePath,
		})
	}
	return results, nil
}

// UpdateAssetType zmienia typ pliku (np. z Image na Texture), ale tylko dla dozwolonych typów.
func (s *AssetService) UpdateAssetType(id int64, newType string) error {
	// 1. Definicja dozwolonych konwersji (Image <-> Texture)
	// Modele, Audio itp. są niezmienne, bo wynikają z rozszerzenia pliku.
	allowedConversion := map[string]bool{
		"image":   true,
		"texture": true,
	}

	if !allowedConversion[newType] {
		return errors.New("invalid target type: only 'image' and 'texture' are allowed")
	}
	asset, err := s.db.GetAssetById(s.ctx, id)
	if err != nil {
		return err
	}

	if !allowedConversion[asset.FileType] {
		return fmt.Errorf("cannot change type for asset of type '%s'", asset.FileType)
	}
	return s.db.UpdateAssetType(s.ctx, database.UpdateAssetTypeParams{
		FileType: newType,
		ID:       id,
	})
}
