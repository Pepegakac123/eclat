package watcher

import (
	"context"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Service struct {
	watcher *fsnotify.Watcher
	logger  *slog.Logger
	ctx     context.Context
	db      database.Querier
}

func NewService(db database.Querier, logger *slog.Logger) (*Service, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Service{
		watcher: w,
		logger:  logger,
		db:      db,
	}, nil
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	go func() {
		s.logger.Info("📂 Initializing folder watchers in background...")
		if err := s.initFolders(); err != nil {
			s.logger.Error("Failed to initialize folders", "error", err)
		} else {
			s.logger.Info("✅ All folders are now being watched")
		}
	}()
	go s.startLoop()
}

func (s *Service) Shutdown() {
	if err := s.watcher.Close(); err != nil {
		s.logger.Error("Failed to close watcher", "error", err)
	}
}

func (s *Service) initFolders() error {
	folders, err := s.db.ListScanFolders(s.ctx)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		err := filepath.WalkDir(folder.Path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				s.logger.Warn("Skipping path due to error", "path", path, "error", err)
				return nil
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			if d.IsDir() {
				s.Watch(path)
			}
			return nil
		})
		if err != nil {
			s.logger.Error("Failed to walk directory", "path", folder, "error", err)
		}
	}
	return nil
}

// Watch dodaje folder do obserwowanych
func (s *Service) Watch(path string) {
	s.logger.Debug("Watching directory", "path", path)
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

			if event.Has(fsnotify.Create) {
				if s.isDir(event.Name) {
					s.logger.Info("New directory detected, adding watcher", "path", event.Name)
					s.Watch(event.Name)
				}
			}
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

func (s *Service) Unwatch(path string) {
	s.logger.Info("Stop watching directory", "path", path)
	if err := s.watcher.Remove(path); err != nil {
		s.logger.Debug("Failed to unwatch path (might not be watched)", "path", path, "error", err)
	}
}
func (s *Service) isDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}
