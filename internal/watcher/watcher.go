package watcher

import (
	"context"
	"eclat/internal/config"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceDuration = 500 * time.Millisecond

type Service struct {
	watcher      *fsnotify.Watcher
	logger       *slog.Logger
	ctx          context.Context
	db           database.Querier
	config       *config.ScannerConfig
	Events       chan string
	watchedPaths map[string]bool
	timers       map[string]*time.Timer
	mu           sync.Mutex
}

func NewService(db database.Querier, logger *slog.Logger) (*Service, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Service{
		watcher:      w,
		logger:       logger,
		db:           db,
		config:       config.NewScannerConfig(),
		Events:       make(chan string, 100),
		timers:       make(map[string]*time.Timer),
		watchedPaths: make(map[string]bool),
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
	close(s.Events)
}

// initFolders pobiera główne foldery z bazy i uruchamia dla nich rekursywne obserwowanie
func (s *Service) initFolders() error {
	folders, err := s.db.ListScanFolders(s.ctx)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if err := s.walkAndWatch(folder.Path); err != nil {
			s.logger.Error("Failed to watch folder tree", "root", folder.Path, "error", err)
		}
	}
	return nil
}

func (s *Service) Watch(path string) {
	s.logger.Info("Adding watcher recursively", "root", path)
	if err := s.walkAndWatch(path); err != nil {
		s.logger.Error("Failed to watch path", "path", path, "error", err)
	}
}

func (s *Service) Unwatch(path string) {
	s.logger.Info("Removing watchers recursively", "root", path)
	s.unwatchRecursive(path)
}

func (s *Service) walkAndWatch(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Debug("Skipping path due to access error", "path", path, "error", err)
			return filepath.SkipDir
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return s.addWatch(path)
		}
		return nil
	})
}

func (s *Service) addWatch(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.watchedPaths[path] {
		return nil
	}

	if err := s.watcher.Add(path); err != nil {
		return err
	}

	s.watchedPaths[path] = true
	// s.logger.Debug("Watching", "path", path) // debug
	return nil
}

func (s *Service) unwatchRecursive(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanRoot := filepath.Clean(root)

	for path := range s.watchedPaths {

		if path == cleanRoot || strings.HasPrefix(path, cleanRoot+string(os.PathSeparator)) {

			if err := s.watcher.Remove(path); err != nil {
				s.logger.Debug("Failed to remove fsnotify watch", "path", path, "error", err)
			}
			delete(s.watchedPaths, path)
		}
	}
}

// --- EVENT LOOP & OPTIMIZATION ---

func (s *Service) startLoop() {
	s.logger.Info("👂 Watcher loop started")
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			// s.logger.Info("🔍 RAW EVENT", "op", event.Op.String(), "path", event.Name) Debug
			if event.Has(fsnotify.Create) && s.isDir(event.Name) {
				s.logger.Info("🆕 New directory detected", "path", event.Name)
				go func(p string) {
					if err := s.walkAndWatch(p); err != nil {
						s.logger.Error("Failed to watch new folder structure", "path", p, "error", err)
					}
				}(event.Name)
				continue
			}

			if !s.isDir(event.Name) {
				if !s.isExtensionAllowed(event.Name) {
					continue
				}
			}

			if s.shouldIgnore(event.Name) {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				s.triggerDebounce(event.Name)
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("Watcher error", "error", err)

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Service) isExtensionAllowed(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(s.config.AllowedExtensions, ext)
}

func (s *Service) triggerDebounce(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, exists := s.timers[path]; exists {
		t.Stop()
	}

	s.timers[path] = time.AfterFunc(debounceDuration, func() {
		s.mu.Lock()
		delete(s.timers, path)
		s.mu.Unlock()

		if _, err := os.Stat(path); err == nil {
			s.logger.Info("🎯 File ready for scan", "path", path)
			select {
			case s.Events <- path:
			default:
				s.logger.Warn("Watcher channel full", "path", path)
			}
		}
	})
}

func (s *Service) isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (s *Service) shouldIgnore(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".") || strings.HasSuffix(path, "~")
}
