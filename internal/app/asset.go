package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// AssetService - Serwis (Zaktualizowany o sysDB)
type AssetService struct {
	ctx           context.Context
	db            database.Querier
	sysDB         *sql.DB
	logger        *slog.Logger
	thumbnailsDir string
}

func NewAssetService(db database.Querier, sysDB *sql.DB, logger *slog.Logger, thumbnailsDir string) *AssetService {
	return &AssetService{
		db:            db,
		sysDB:         sysDB,
		logger:        logger,
		thumbnailsDir: thumbnailsDir,
	}
}

func (s *AssetService) Startup(ctx context.Context) {
	s.ctx = ctx
	// Migracja ścieżek miniaturek przy starcie
	go func() {
		err := s.MigrateThumbnailPaths()
		if err != nil {
			s.logger.Error("Failed to migrate thumbnail paths", "error", err)
		}
	}()
}

// MigrateThumbnailPaths konwertuje bezwzględne ścieżki systemowe na ścieżki relatywne /thumbnails/
func (s *AssetService) MigrateThumbnailPaths() error {
	ctx := context.Background()
	s.logger.Info("Starting thumbnail path migration...")

	// Pobierz wszystkie assety z potencjalnie błędnymi ścieżkami
	rows, err := s.sysDB.QueryContext(ctx, `
		SELECT id, thumbnail_path FROM assets
		WHERE thumbnail_path LIKE '/%'
		AND thumbnail_path NOT LIKE '/thumbnails/%'
		AND thumbnail_path NOT LIKE '/placeholders/%'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type update struct {
		id   int64
		path string
	}
	var updates []update

	for rows.Next() {
		var id int64
		var oldPath string
		if err := rows.Scan(&id, &oldPath); err != nil {
			continue
		}

		// Wyciągnij samą nazwę pliku
		filename := filepath.Base(oldPath)
		newPath := "/thumbnails/" + filename
		updates = append(updates, update{id, newPath})
	}

	if len(updates) == 0 {
		s.logger.Info("No thumbnails need migration")
		return nil
	}

	s.logger.Info("Migrating thumbnails", "count", len(updates))

	tx, err := s.sysDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE assets SET thumbnail_path = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range updates {
		_, err := stmt.ExecContext(ctx, u.path, u.id)
		if err != nil {
			s.logger.Error("Failed to update thumbnail path", "id", u.id, "error", err)
		}
	}

	return tx.Commit()
}

// ==========================================
// Data Transfer Objects (DTOs) for Wails
// ==========================================

// AssetMaterialSet - Uproszczona struktura zestawu dla assetu
type AssetMaterialSet struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CustomColor string `json:"customColor"`
}

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

	// Nowe pola (Naprawione błędy undefined)
	ImageWidth    int64  `json:"imageWidth"`
	ImageHeight   int64  `json:"imageHeight"`
	FileExtension string `json:"fileExtension"`

	// Metadane
	Rating      int64  `json:"rating"`
	IsFavorite  bool   `json:"isFavorite"`
	Description string `json:"description"`
	IsDeleted   bool   `json:"isDeleted"`
	IsHidden    bool   `json:"isHidden"`
	BitDepth    int64  `json:"bitDepth"` // Added
	FileHash    string `json:"fileHash"` // Added

	// Relacje
	GroupID       *string            `json:"groupId"`
	Tags          []string           `json:"tags"`
	MaterialSets  []AssetMaterialSet `json:"materialSets"`
	DominantColor string             `json:"dominantColor"`
}
type AssetSibling struct {
	ID       int64  `json:"id"`
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
}

// AssetQueryFilters - Filtry z Frontendu
type AssetQueryFilters struct {
	Page         int      `json:"page"`
	PageSize     int      `json:"pageSize"`
	Query        string   `json:"searchQuery"`
	Tags         []string `json:"tags"`
	MatchAllTags bool     `json:"matchAllTags"`
	FileTypes    []string `json:"fileTypes"`
	Colors       []string `json:"colors"`

	RatingRange   []int `json:"ratingRange"`   // [min, max]
	WidthRange    []int `json:"widthRange"`    // [min, max]
	HeightRange   []int `json:"heightRange"`   // [min, max]
	FileSizeRange []int `json:"fileSizeRange"` // [min, max] MB

	DateRange struct {
		From *string `json:"from"`
		To   *string `json:"to"`
	} `json:"dateRange"`

	HasAlpha          *bool  `json:"hasAlpha"`
	OnlyFavorites     bool   `json:"onlyFavorites"`
	OnlyUncategorized bool   `json:"onlyUncategorized"`
	IsDeleted         bool   `json:"isDeleted"`
	IsHidden          bool   `json:"isHidden"`
	CollectionID      *int64 `json:"collectionId"`

	SortOption string `json:"sortOption"`
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

// GetAssets to główna metoda galerii, obsługująca dynamiczne filtrowanie.
func (s *AssetService) GetAssets(filters AssetQueryFilters) (*PagedAssetResult, error) {
	// Inicjalizacja Buildera dla SQLite
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Question)

	// Bazowe zapytanie
	base := psql.Select(
		"a.id", "a.file_name", "a.file_path", "a.file_type", "a.file_size",
		"a.thumbnail_path", "a.date_added", "a.last_modified",
		"a.image_width", "a.image_height", "a.dominant_color",
		"a.rating", "a.is_favorite", "a.is_deleted", "a.is_hidden", "a.group_id",
		"a.has_alpha_channel", "a.bit_depth", "a.file_hash", "a.description",
	).From("assets a")

	// Jeśli nie szukamy w koszu, filtrujemy po folderach
	if !filters.IsDeleted {
		base = base.Join("scan_folders f ON a.scan_folder_id = f.id").
			Where(sq.Eq{"f.is_deleted": 0}).
			Where(sq.Eq{"f.is_active": 1})
	}

	// Filtrowanie Podstawowe
	base = base.Where(sq.Eq{"a.is_deleted": filters.IsDeleted})
	if !filters.IsDeleted {
		base = base.Where(sq.Eq{"a.is_hidden": filters.IsHidden})
	}

	if filters.OnlyFavorites {
		base = base.Where(sq.Eq{"a.is_favorite": 1})
	}

	// Filtrowanie po Kolekcji
	if filters.CollectionID != nil {
		base = base.Join("asset_material_sets msa ON a.id = msa.asset_id").
			Where(sq.Eq{"msa.material_set_id": *filters.CollectionID})
	}

	// Wyszukiwanie Tekstowe
	if filters.Query != "" {
		like := "%" + filters.Query + "%"
		base = base.Where(sq.Or{
			sq.Like{"a.file_name": like},
			sq.Like{"a.file_path": like},
		})
	}

	// === ZAKRESY (Poprawione Gte -> GtOrEq, Lte -> LtOrEq) ===

	// Rating
	if len(filters.RatingRange) == 2 {
		if filters.RatingRange[0] > 0 {
			base = base.Where(sq.GtOrEq{"a.rating": filters.RatingRange[0]})
		}
		if filters.RatingRange[1] < 5 {
			base = base.Where(sq.LtOrEq{"a.rating": filters.RatingRange[1]})
		}
	}

	// File Size (MB -> Bytes)
	if len(filters.FileSizeRange) == 2 {
		if filters.FileSizeRange[0] > 0 {
			minBytes := int64(filters.FileSizeRange[0]) * 1024 * 1024
			base = base.Where(sq.GtOrEq{"a.file_size": minBytes})
		}
		if filters.FileSizeRange[1] < 4096 { // 4GB max in UI usually
			maxBytes := int64(filters.FileSizeRange[1]) * 1024 * 1024
			base = base.Where(sq.LtOrEq{"a.file_size": maxBytes})
		}
	}

	// Dimensions - FIXED: Only apply if min > 0 OR max < DEFAULT_MAX.
	// This ensures that NULL values (models) are NOT filtered out when range is 0-MAX.
	const defaultMaxDim = 8160 // Matches frontend UI_CONFIG.GALLERY.FilterOptions.MAX_DIMENSION
	if len(filters.WidthRange) == 2 {
		if filters.WidthRange[0] > 0 {
			base = base.Where(sq.GtOrEq{"a.image_width": filters.WidthRange[0]})
		}
		if filters.WidthRange[1] > 0 && filters.WidthRange[1] < defaultMaxDim {
			base = base.Where(sq.LtOrEq{"a.image_width": filters.WidthRange[1]})
		}
	}

	if len(filters.HeightRange) == 2 {
		if filters.HeightRange[0] > 0 {
			base = base.Where(sq.GtOrEq{"a.image_height": filters.HeightRange[0]})
		}
		if filters.HeightRange[1] > 0 && filters.HeightRange[1] < defaultMaxDim {
			base = base.Where(sq.LtOrEq{"a.image_height": filters.HeightRange[1]})
		}
	}

	// Daty
	if filters.DateRange.From != nil && *filters.DateRange.From != "" {
		base = base.Where(sq.GtOrEq{"a.date_added": *filters.DateRange.From})
	}
	if filters.DateRange.To != nil && *filters.DateRange.To != "" {
		base = base.Where(sq.LtOrEq{"a.date_added": *filters.DateRange.To})
	}

	// Typy plików
	if len(filters.FileTypes) > 0 {
		base = base.Where(sq.Eq{"a.file_type": filters.FileTypes})
	}

	// Kolory
	if len(filters.Colors) > 0 {
		base = base.Where(sq.Eq{"a.dominant_color": filters.Colors})
	}

	// Alpha Channel
	if filters.HasAlpha != nil {
		val := 0
		if *filters.HasAlpha {
			val = 1
		}
		base = base.Where(sq.Eq{"a.has_alpha_channel": val})
	}

	// Tagi (Subquery)
	if len(filters.Tags) > 0 {
		tagSubQ := sq.Select("at.asset_id").
			From("asset_tags at").
			Join("tags t ON at.tag_id = t.id").
			Where(sq.Eq{"t.name": filters.Tags})

		if filters.MatchAllTags {
			tagSubQ = tagSubQ.GroupBy("at.asset_id").
				Having(sq.Eq{"COUNT(DISTINCT t.id)": len(filters.Tags)})
		}

		// Squirrel wymaga ręcznego SQL dla podzapytania w IN
		subSql, subArgs, err := tagSubQ.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build tag subquery: %w", err)
		}
		base = base.Where(fmt.Sprintf("a.id IN (%s)", subSql), subArgs...)
	} else if filters.OnlyUncategorized {
		// Assets with NO tags
		base = base.Where("NOT EXISTS (SELECT 1 FROM asset_tags at WHERE at.asset_id = a.id)")
	}

	// ==========================================
	// COUNT (Używamy s.sysDB)
	// ==========================================

	sqlBase, argsBase, err := base.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build base sql: %w", err)
	}

	fromIndex := strings.Index(strings.ToUpper(sqlBase), "FROM")
	if fromIndex == -1 {
		return nil, fmt.Errorf("invalid sql generated (no FROM clause)")
	}
	sqlCount := "SELECT COUNT(*) " + sqlBase[fromIndex:]

	var totalCount int
	err = s.sysDB.QueryRowContext(context.Background(), sqlCount, argsBase...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	// ==========================================
	// Sortowanie i Paginacja
	// ==========================================

	sortCol := "a.date_added"
	switch strings.ToLower(filters.SortOption) {
	case "filename":
		sortCol = "a.file_name"
	case "filesize":
		sortCol = "a.file_size"
	case "lastmodified":
		sortCol = "a.last_modified"
	case "rating":
		sortCol = "a.rating"
	case "dateadded":
		sortCol = "a.date_added"
	}

	sortDir := "DESC"
	if !filters.SortDesc {
		sortDir = "ASC"
	}
	base = base.OrderBy(fmt.Sprintf("%s %s", sortCol, sortDir))

	offset := (filters.Page - 1) * filters.PageSize
	if offset < 0 {
		offset = 0
	}
	base = base.Limit(uint64(filters.PageSize)).Offset(uint64(offset))

	// ==========================================
	// Wykonanie i Mapowanie
	// ==========================================

	finalSQL, finalArgs, err := base.ToSql()
	if err != nil {
		return nil, err
	}

	// POPRAWKA: Używamy s.sysDB.QueryContext
	rows, err := s.sysDB.QueryContext(context.Background(), finalSQL, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AssetDetails

	for rows.Next() {
		var a AssetDetails

		// Zmienne tymczasowe dla typów nullable
		var thumbPath sql.NullString
		var domColor sql.NullString
		var imgW, imgH sql.NullInt64
		var hasAlpha sql.NullBool
		var dateAddedStr string // SQLite TEXT
		var lastModTime time.Time
		var groupId sql.NullString
		var bitDepth sql.NullInt64
		var fileHash sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&a.ID, &a.FileName, &a.FilePath, &a.FileType, &a.FileSize,
			&thumbPath, &dateAddedStr, &lastModTime,
			&imgW, &imgH, &domColor,
			&a.Rating, &a.IsFavorite, &a.IsDeleted, &a.IsHidden, &groupId,
			&hasAlpha, &bitDepth, &fileHash, &description,
		)
		if err != nil {
			return nil, err
		}

		// Mapowanie wartości
		if groupId.Valid && groupId.String != "" {
			val := groupId.String
			a.GroupID = &val
		} else {
			a.GroupID = nil
		}

		if bitDepth.Valid {
			a.BitDepth = bitDepth.Int64
		}
		if fileHash.Valid {
			a.FileHash = fileHash.String
		}
		if description.Valid {
			a.Description = description.String
		}

		if thumbPath.Valid {
			a.ThumbnailPath = thumbPath.String
		} else {
			a.ThumbnailPath = ""
		}

		if domColor.Valid {
			a.DominantColor = domColor.String
		} else {
			a.DominantColor = ""
		}

		a.LastModified = lastModTime

		if imgW.Valid {
			a.ImageWidth = imgW.Int64
		} else {
			a.ImageWidth = 0
		}

		if imgH.Valid {
			a.ImageHeight = imgH.Int64
		} else {
			a.ImageHeight = 0
		}

		a.FileExtension = strings.ToLower(filepath.Ext(a.FileName))

		if groupId.Valid {
			val := groupId.String
			a.GroupID = &val
		}

		// Parsowanie daty
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", dateAddedStr); err == nil {
			a.DateAdded = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, dateAddedStr); err == nil {
			a.DateAdded = parsedTime
		} else {
			a.DateAdded = time.Now()
		}

		results = append(results, a)
	}

	if results == nil {
		results = []AssetDetails{}
	}

	if len(results) > 0 {
		s.logger.Debug("Returning assets", "count", len(results), "firstThumb", results[0].ThumbnailPath)
	}

	return &PagedAssetResult{
		Items:      results,
		TotalCount: totalCount,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
	}, nil
}

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

	// 3. Pobierz Material Sets (RAW SQL dla bezpieczeństwa)
	// Uwaga: Zakładam, że nazwy tabel są poprawne (material_sets, material_set_assets)
	var materialSets []AssetMaterialSet
	msRows, err := s.sysDB.QueryContext(ctx, `
			SELECT ms.id, ms.name, ms.custom_color
			FROM material_sets ms
			JOIN asset_material_sets msa ON ms.id = msa.material_set_id
			WHERE msa.asset_id = ?
		`, asset.ID)
	if err == nil {
		defer msRows.Close()
		for msRows.Next() {
			var ms AssetMaterialSet
			var customColor sql.NullString
			if err := msRows.Scan(&ms.ID, &ms.Name, &customColor); err == nil {
				if customColor.Valid {
					ms.CustomColor = customColor.String
				}
				materialSets = append(materialSets, ms)
			}
		}
	} else {
		// Log error but don't fail the whole request
		s.logger.Error("Failed to fetch material sets", "error", err)
	}

	// 4. Mapowanie
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
		GroupID:       &asset.GroupID,
		Tags:          tagNames,
		DominantColor: asset.DominantColor.String,

		// New fields
		BitDepth:     asset.BitDepth.Int64,
		FileHash:     asset.FileHash.String,
		MaterialSets: materialSets,

		// Calculated
		ImageWidth:    asset.ImageWidth.Int64,
		ImageHeight:   asset.ImageHeight.Int64,
		FileExtension: strings.ToLower(filepath.Ext(asset.FileName)),
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
		switch v := stats.LastScan.(type) {
		case time.Time:
			lastScan = &v
		case []byte:
			// SQLite driver sometimes returns []byte for strings
			sVal := string(v)
			lastScan = parseTime(sVal, s.logger)
		case string:
			lastScan = parseTime(v, s.logger)
		default:
			s.logger.Warn("Unknown type for LastScan", "type", fmt.Sprintf("%T", v), "value", v)
		}
	}

	return &LibraryStats{
		TotalAssets: stats.TotalCount,
		TotalSize:   stats.TotalSize,
		LastScan:    lastScan,
	}, nil
}

func parseTime(v string, logger *slog.Logger) *time.Time {
	// Strip monotonic clock info if present (e.g., " ... m=+123.456")
	if idx := strings.Index(v, " m="); idx != -1 {
		v = v[:idx]
	}

	formats := []string{
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, v); err == nil {
			return &t
		}
	}
	logger.Warn("Could not parse LastScan string", "value", v)
	return nil
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
	for _, id := range ids {
		// 1. Pobierz dane assetu, aby poznać ścieżkę pliku
		asset, err := s.db.GetAssetById(s.ctx, id)
		if err != nil {
			s.logger.Error("Failed to get asset for deletion", "id", id, "error", err)
			return fmt.Errorf("failed to retrieve asset %d: %w", id, err)
		}

		// 2. Usuń plik z dysku
		if err := os.Remove(asset.FilePath); err != nil && !os.IsNotExist(err) {
			s.logger.Error("Failed to delete file from disk", "path", asset.FilePath, "error", err)
			return fmt.Errorf("failed to delete file %s: %w", asset.FilePath, err)
		}

		// 3. Usuń miniaturkę, jeśli istnieje i jest wygenerowana (nie jest placeholderem)
		if asset.ThumbnailPath != "" && strings.HasPrefix(asset.ThumbnailPath, "/thumbnails/") {
			thumbName := filepath.Base(asset.ThumbnailPath)
			thumbFullPath := filepath.Join(s.thumbnailsDir, thumbName)
			if err := os.Remove(thumbFullPath); err != nil && !os.IsNotExist(err) {
				s.logger.Warn("Failed to delete thumbnail", "path", thumbFullPath, "error", err)
			}
		}

		// 4. Usuń wpis z bazy danych
		if err := s.db.DeleteAssetPermanent(s.ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// RenameAsset zmienia nazwę pliku na dysku i w bazie danych.
func (s *AssetService) RenameAsset(id int64, newName string) error {
	// 1. Walidacja nowej nazwy (prosta)
	if newName == "" {
		return errors.New("new name cannot be empty")
	}
	if strings.ContainsAny(newName, `/\:*?"<>|`) {
		return errors.New("invalid characters in filename")
	}

	// 2. Pobierz asset
	asset, err := s.db.GetAssetById(s.ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get asset: %w", err)
	}

	originalExt := filepath.Ext(asset.FileName)
	// Upewnij się, że nowa nazwa ma to samo rozszerzenie
	if !strings.HasSuffix(strings.ToLower(newName), strings.ToLower(originalExt)) {
		// Jeśli użytkownik usunął rozszerzenie, dodaj je z powrotem
		// Jeśli zmienił na inne - to też zostanie nadpisane oryginalnym (lub dodane)
		// Przyjmijmy strategię: jeśli nie ma suffixu, dodajemy.
		newName = newName + originalExt
	} else {
        // Case sensitive check? Windows is case insensitive, Linux sensitive.
        // Let's force the original extension case if possible, or just accept what user gave if it matches.
        // For safety, let's reconstruct the name with original extension to be sure.
        baseName := newName[:len(newName)-len(originalExt)]
        newName = baseName + originalExt
    }

	// 3. Przygotuj ścieżki
	dir := filepath.Dir(asset.FilePath)
	newPath := filepath.Join(dir, newName)

	// 4. Sprawdź czy plik docelowy istnieje
	if _, err := os.Stat(newPath); err == nil {
		return errors.New("file with this name already exists")
	}

	// 5. Zmień nazwę pliku na dysku
	if err := os.Rename(asset.FilePath, newPath); err != nil {
		return fmt.Errorf("failed to rename file on disk: %w", err)
	}

	// 6. Zaktualizuj bazę danych
	// Nie zmieniamy miniatury, bo jej nazwa jest generowana dynamicznie/haszowana i nie zależy od nazwy pliku (według instrukcji).

	params := database.RenameAssetParams{
		FileName: newName,
		FilePath: newPath,
		ID:       id,
	}

	_, err = s.db.RenameAsset(s.ctx, params)
	if err != nil {
		// Rollback rename on disk?
		// _ = os.Rename(newPath, asset.FilePath)
		return fmt.Errorf("failed to update asset in db: %w", err)
	}

	return nil
}

// GetAvailableColors zwraca listę wszystkich unikalnych kolorów dominujących z bazy danych.
func (s *AssetService) GetAvailableColors() ([]string, error) {
	nullColors, err := s.db.GetAllColors(s.ctx)
	if err != nil {
		return nil, err
	}

	var colors []string
	for _, c := range nullColors {
		if c.Valid && c.String != "" {
			colors = append(colors, c.String)
		}
	}
	return colors, nil
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

// UpdateTags aktualizuje listę tagów dla assetu.
func (s *AssetService) UpdateTags(assetId int64, tags []string) error {
	tx, err := s.sysDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := database.New(s.sysDB).WithTx(tx)

	// 1. Clear existing tags
	if err := qtx.ClearTagsForAsset(s.ctx, assetId); err != nil {
		return err
	}

	// 2. Add new tags
	for _, tagName := range tags {
		// Ensure tag exists
		tag, err := qtx.CreateTag(s.ctx, tagName)
		if err != nil {
			return err
		}
		// Link tag
		err = qtx.AddTagToAsset(s.ctx, database.AddTagToAssetParams{
			AssetID: assetId,
			TagID:   tag.ID,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddAssetToMaterialSet dodaje asset do kolekcji.
func (s *AssetService) AddAssetToMaterialSet(setId int64, assetId int64) error {
	return s.db.AddAssetToMaterialSet(s.ctx, database.AddAssetToMaterialSetParams{
		MaterialSetID: setId,
		AssetID:       assetId,
	})
}

// RemoveAssetFromMaterialSet usuwa asset z kolekcji.
func (s *AssetService) RemoveAssetFromMaterialSet(setId int64, assetId int64) error {
	return s.db.RemoveAssetFromMaterialSet(s.ctx, database.RemoveAssetFromMaterialSetParams{
		MaterialSetID: setId,
		AssetID:       assetId,
	})
}

// GetThumbnailData returns the thumbnail image as a base64 data URL.
// This is a workaround for issues serving dynamic assets via Wails Handler in some Dev environments.
func (s *AssetService) GetThumbnailData(assetId int64) (string, error) {
	asset, err := s.db.GetAssetById(context.Background(), assetId)
	if err != nil {
		return "", err
	}

	if asset.ThumbnailPath == "" {
		return "", errors.New("no thumbnail path")
	}

	// If it's a placeholder, we return the path as is (frontend will handle it)
	if strings.HasPrefix(asset.ThumbnailPath, "/placeholders/") {
		return asset.ThumbnailPath, nil
	}

	// Resolve absolute path
	filename := filepath.Base(asset.ThumbnailPath)
	fullPath := filepath.Join(s.thumbnailsDir, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read thumbnail: %w", err)
	}

	// Determine MIME type (usually webp)
	mimeType := "image/webp"
	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		mimeType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") || strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		mimeType = "image/jpeg"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}
