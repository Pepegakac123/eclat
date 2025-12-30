package app

import (
	"context"
	"eclat/internal/database"
	"log/slog"
)

type TagService struct {
	ctx    context.Context
	db     database.Querier
	logger *slog.Logger
}

func NewTagService(db database.Querier, logger *slog.Logger) *TagService {
	return &TagService{
		db:     db,
		logger: logger,
	}
}

func (s *TagService) Startup(ctx context.Context) {
	s.ctx = ctx
}

type Tag struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	AssetCount int64  `json:"assetCount"`
}

// GetAll returns all tags with asset counts.
func (s *TagService) GetAll() ([]Tag, error) {
	rows, err := s.db.ListTags(s.ctx)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	for _, r := range rows {
		tags = append(tags, Tag{
			ID:         r.ID,
			Name:       r.Name,
			AssetCount: r.AssetCount,
		})
	}
	if tags == nil {
		tags = []Tag{}
	}
	return tags, nil
}
