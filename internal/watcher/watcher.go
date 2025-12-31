package watcher

import (
	"context"
	"eclat/internal/config"
	"eclat/internal/database"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounceDuration defines the time window to group multiple file events into a single action.
const debounceDuration = 500 * time.Millisecond

// Service implements a file system watcher that monitors directories for changes.
// It reports file creations, modifications, and deletions to the scanner service.
type Service struct {
	watcher      *fsnotify.Watcher
	logger       *slog.Logger
	ctx          context.Context
	db           database.Querier
	config       *config.ScannerConfig  // Shared configuration for allowed extensions
	Events       chan string            // Channel to emit file paths that need scanning
	watchedPaths map[string]bool        // Set of currently watched directories
	timers       map[string]*time.Timer // Debounce timers for active file events
	mu           sync.Mutex
	shutdownOnce sync.Once
}

// NewService creates a new Watcher Service instance.
// It requires a database connection to load initial folders and a configuration for extension filtering.
func NewService(db database.Querier, logger *slog.Logger, cfg *config.ScannerConfig) (*Service, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Service{
		watcher:      w,
		logger:       logger,
		db:           db,
		config:       cfg,
		Events:       make(chan string, 1000),
		timers:       make(map[string]*time.Timer),
		watchedPaths: make(map[string]bool),
	}, nil
}

// Startup initializes the watcher service.
// It loads configured scan folders from the database and starts the event processing loop.
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

// Shutdown gracefully stops the watcher, closes the fsnotify instance, and cleans up resources.
func (s *Service) Shutdown() {
	s.shutdownOnce.Do(func() {
		s.logger.Info("🛑 Shutting down Watcher service...")
		if err := s.watcher.Close(); err != nil {
			s.logger.Error("Failed to close fsnotify watcher", "error", err)
		}
		s.mu.Lock()
		for _, t := range s.timers {
			t.Stop()
		}
		s.mu.Unlock()

		close(s.Events)
	})
}

// initFolders loads all active scan folders from the database and adds them to the watcher.
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

// Watch adds a new directory (and its subdirectories) to the watcher.
func (s *Service) Watch(path string) {
	s.logger.Info("Adding watcher recursively", "root", path)
	if err := s.walkAndWatch(path); err != nil {
		s.logger.Error("Failed to watch path", "path", path, "error", err)
	}
}

// Unwatch removes a directory (and its subdirectories) from the watcher.
func (s *Service) Unwatch(path string) {
	s.logger.Info("Removing watchers recursively", "root", path)
	s.unwatchRecursive(path)
}

// walkAndWatch recursively walks a directory tree and adds each directory to the watcher.
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

// addWatch adds a single directory to the fsnotify watcher.
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
	return nil
}

// unwatchRecursive removes a directory and all its nested watched paths from the watcher.
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

// startLoop processes events from the fsnotify channel.
// It handles new directory detection, extension filtering, and debouncing of file events.
func (s *Service) startLoop() {
	s.logger.Info("👂 Watcher loop started")
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
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

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
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
	// Delegates to the thread-safe method from config
	return s.config.IsExtensionAllowed(path)
}

// triggerDebounce starts or resets a timer for a specific file path.
// When the timer expires, the file is sent for scanning. This prevents multiple scans for a single file operation.
func (s *Service) triggerDebounce(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.ctx.Done():
		return
	default:
	}

	if t, exists := s.timers[path]; exists {
		t.Stop()
	}

	s.timers[path] = time.AfterFunc(debounceDuration, func() {
		s.mu.Lock()
		// Check if timer is still valid (wasn't cleared by Shutdown)
		if _, ok := s.timers[path]; !ok {
			s.mu.Unlock()
			return
		}
		delete(s.timers, path)
		s.mu.Unlock()

		_, err := os.Stat(path)

		if err == nil {
			s.logger.Info(" File ready for scan", "path", path)
			s.sendEvent(path)
			return
		}

		if os.IsNotExist(err) {
			s.logger.Info("File deletion detected", "path", path)
			s.sendEvent(path)
			return
		}
	})
}

// sendEvent puts a file path into the events channel for the scanner to consume.
func (s *Service) sendEvent(path string) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Warn("Dropped event due to channel close", "path", path)
		}
	}()

	select {
	case s.Events <- path:
	default:
		s.logger.Warn("Watcher channel full", "path", path)
	}
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
