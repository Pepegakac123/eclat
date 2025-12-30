package app

import (
	"context"
	"database/sql"
	"eclat/internal/database"
	"log/slog"
	"time"
)

type MaterialSetService struct {
	ctx    context.Context
	db     database.Querier
	logger *slog.Logger
}

func NewMaterialSetService(db database.Querier, logger *slog.Logger) *MaterialSetService {
	return &MaterialSetService{
		db:     db,
		logger: logger,
	}
}

func (s *MaterialSetService) Startup(ctx context.Context) {
	s.ctx = ctx
}

type MaterialSet struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	CoverAssetID   *int64    `json:"coverAssetId"`
	CustomCoverUrl *string   `json:"customCoverUrl"`
	CustomColor    *string   `json:"customColor"`
	DateAdded      time.Time `json:"dateAdded"`
	LastModified   time.Time `json:"lastModified"`
	TotalAssets    int64     `json:"totalAssets"`
}

type CreateMaterialSetRequest struct {
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	CoverAssetID   *int64  `json:"coverAssetId"`
	CustomCoverUrl *string `json:"customCoverUrl"`
	CustomColor    *string `json:"customColor"`
}

// GetAll returns all material sets.
func (s *MaterialSetService) GetAll() ([]MaterialSet, error) {
	rows, err := s.db.ListMaterialSets(s.ctx)
	if err != nil {
		return nil, err
	}

	var results []MaterialSet
	for _, r := range rows {
		var desc, customUrl, customColor *string
		if r.Description.Valid {
			val := r.Description.String
			desc = &val
		}
		if r.CustomCoverUrl.Valid {
			val := r.CustomCoverUrl.String
			customUrl = &val
		}
		if r.CustomColor.Valid {
			val := r.CustomColor.String
			customColor = &val
		}

		var coverId *int64
		if r.CoverAssetID.Valid {
			val := r.CoverAssetID.Int64
			coverId = &val
		}

		results = append(results, MaterialSet{
			ID:             r.ID,
			Name:           r.Name,
			Description:    desc,
			CoverAssetID:   coverId,
			CustomCoverUrl: customUrl,
			CustomColor:    customColor,
			DateAdded:      r.DateAdded,
			LastModified:   r.LastModified,
			TotalAssets:    r.TotalAssets,
		})
	}
	if results == nil {
		results = []MaterialSet{}
	}
	return results, nil
}

// Create creates a new material set.
func (s *MaterialSetService) Create(req CreateMaterialSetRequest) (*MaterialSet, error) {
	params := database.CreateMaterialSetParams{
		Name: req.Name,
		Description: sql.NullString{
			String: getString(req.Description),
			Valid:  req.Description != nil,
		},
		CoverAssetID: sql.NullInt64{
			Int64: getInt64(req.CoverAssetID),
			Valid: req.CoverAssetID != nil,
		},
		CustomCoverUrl: sql.NullString{
			String: getString(req.CustomCoverUrl),
			Valid:  req.CustomCoverUrl != nil,
		},
		CustomColor: sql.NullString{
			String: getString(req.CustomColor),
			Valid:  req.CustomColor != nil,
		},
	}

	ms, err := s.db.CreateMaterialSet(s.ctx, params)
	if err != nil {
		return nil, err
	}

	return s.GetById(ms.ID)
}

// Update updates a material set.
func (s *MaterialSetService) Update(id int64, req CreateMaterialSetRequest) (*MaterialSet, error) {
	params := database.UpdateMaterialSetParams{
		ID:   id,
		Name: req.Name,
		Description: sql.NullString{
			String: getString(req.Description),
			Valid:  req.Description != nil,
		},
		CoverAssetID: sql.NullInt64{
			Int64: getInt64(req.CoverAssetID),
			Valid: req.CoverAssetID != nil,
		},
		CustomCoverUrl: sql.NullString{
			String: getString(req.CustomCoverUrl),
			Valid:  req.CustomCoverUrl != nil,
		},
		CustomColor: sql.NullString{
			String: getString(req.CustomColor),
			Valid:  req.CustomColor != nil,
		},
	}

	if err := s.db.UpdateMaterialSet(s.ctx, params); err != nil {
		return nil, err
	}
	return s.GetById(id)
}

// Delete deletes a material set.
func (s *MaterialSetService) Delete(id int64) error {
	return s.db.DeleteMaterialSet(s.ctx, id)
}

// GetById gets a material set by ID.
func (s *MaterialSetService) GetById(id int64) (*MaterialSet, error) {
	ms, err := s.db.GetMaterialSetById(s.ctx, id)
	if err != nil {
		return nil, err
	}

	var desc, customUrl, customColor *string
	if ms.Description.Valid {
		val := ms.Description.String
		desc = &val
	}
	if ms.CustomCoverUrl.Valid {
		val := ms.CustomCoverUrl.String
		customUrl = &val
	}
	if ms.CustomColor.Valid {
		val := ms.CustomColor.String
		customColor = &val
	}

	var coverId *int64
	if ms.CoverAssetID.Valid {
		val := ms.CoverAssetID.Int64
		coverId = &val
	}

	return &MaterialSet{
		ID:             ms.ID,
		Name:           ms.Name,
		Description:    desc,
		CoverAssetID:   coverId,
		CustomCoverUrl: customUrl,
		CustomColor:    customColor,
		DateAdded:      ms.DateAdded,
		LastModified:   ms.LastModified,
		TotalAssets:    0, // Default to 0 as query doesn't include count
	}, nil
}

// Helpers
func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func getInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
