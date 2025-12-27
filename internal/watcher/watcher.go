package watcher

import (
	"context"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type Service struct {
	watcher *fsnotify.Watcher
	logger  *slog.Logger
	ctx     context.Context
}

func NewService(logger *slog.Logger) (*Service, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Service{
		watcher: w,
		logger:  logger,
	}, nil
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	go s.startLoop()
}

func (s *Service) Shutdown() {
	if err := s.watcher.Close(); err != nil {
		s.logger.Error("Failed to close watcher", "error", err)
	}
}

// Watch dodaje folder do obserwowanych
func (s *Service) Watch(path string) {
	s.logger.Info("Watching directory", "path", path)
	if err := s.watcher.Add(path); err != nil {
		s.logger.Error("Failed to watch path", "path", path, "error", err)
	}
}

// startLoop to serce - pętla nasłuchująca
func (s *Service) startLoop() {
	s.logger.Info("Watcher loop started")
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			// Na razie tylko logujemy - żeby zobaczyć czy działa
			s.logger.Info("🔔 File Event", "op", event.Op.String(), "name", event.Name)

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("Watcher error", "error", err)

		case <-s.ctx.Done():
			s.logger.Info("Watcher stopped by context")
			return
		}
	}
}
