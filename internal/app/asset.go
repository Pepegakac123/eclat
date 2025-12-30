package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// AssetService - Serwis (Zaktualizowany o sysDB)
type AssetService struct {
	ctx    context.Context
	db     database.Querier
	sysDB  *sql.DB
	logger *slog.Logger
}

func NewAssetService(db database.Querier, sysDB *sql.DB, logger *slog.Logger) *AssetService {
	return &AssetService{
		db:     db,
		sysDB:  sysDB,
		logger: logger,
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

	HasAlpha      *bool  `json:"hasAlpha"`
	OnlyFavorites bool   `json:"onlyFavorites"`
	IsDeleted     bool   `json:"isDeleted"`
	IsHidden      bool   `json:"isHidden"`
	CollectionID  *int64 `json:"collectionId"`

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
		"a.has_alpha_channel",
	).From("assets a")

	// Jeśli nie szukamy w koszu, filtrujemy po folderach
	if !filters.IsDeleted {
		base = base.Join("scan_folders f ON a.scan_folder_id = f.id").
			Where(sq.Eq{"f.is_deleted": 0}).
			Where(sq.Eq{"f.is_active": 1})
	}

	// Filtrowanie Podstawowe
	base = base.Where(sq.Eq{"a.is_deleted": filters.IsDeleted})
	base = base.Where(sq.Eq{"a.is_hidden": filters.IsHidden})

	if filters.OnlyFavorites {
		base = base.Where(sq.Eq{"a.is_favorite": 1})
	}

	// Filtrowanie po Kolekcji
	if filters.CollectionID != nil {
		base = base.Join("material_set_assets msa ON a.id = msa.asset_id").
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
		base = base.Where(sq.GtOrEq{"a.rating": filters.RatingRange[0]})
		base = base.Where(sq.LtOrEq{"a.rating": filters.RatingRange[1]})
	}

	// File Size (MB -> Bytes)
	if len(filters.FileSizeRange) == 2 {
		minBytes := int64(filters.FileSizeRange[0]) * 1024 * 1024
		maxBytes := int64(filters.FileSizeRange[1]) * 1024 * 1024
		base = base.Where(sq.GtOrEq{"a.file_size": minBytes})
		base = base.Where(sq.LtOrEq{"a.file_size": maxBytes})
	}

	// Width
	if len(filters.WidthRange) == 2 && filters.WidthRange[1] > 0 {
		base = base.Where(sq.GtOrEq{"a.image_width": filters.WidthRange[0]})
		base = base.Where(sq.LtOrEq{"a.image_width": filters.WidthRange[1]})
	}

	// Height
	if len(filters.HeightRange) == 2 && filters.HeightRange[1] > 0 {
		base = base.Where(sq.GtOrEq{"a.image_height": filters.HeightRange[0]})
		base = base.Where(sq.LtOrEq{"a.image_height": filters.HeightRange[1]})
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

		err := rows.Scan(
			&a.ID, &a.FileName, &a.FilePath, &a.FileType, &a.FileSize,
			&thumbPath, &dateAddedStr, &lastModTime,
			&imgW, &imgH, &domColor,
			&a.Rating, &a.IsFavorite, &a.IsDeleted, &a.IsHidden, &groupId,
			&hasAlpha,
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
