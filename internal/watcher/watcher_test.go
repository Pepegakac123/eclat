package watcher

import (
	"context"
	"eclat/internal/config" // <--- Nowy import
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// 1. TEST: DEBOUNCING
// Użytkownik zapisuje plik 5 razy w ułamku sekundy. Oczekujemy tylko 1 zdarzenia.
func TestWatcher_Integrity_Debouncing(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	svc.Startup(ctx)
	time.Sleep(100 * time.Millisecond)

	filePath := filepath.Join(root, "spam.png")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			createDummyFile(t, filePath)
			time.Sleep(50 * time.Millisecond)
		}
	}()
	wg.Wait()

	// Oczekujemy dokładnie jednego zdarzenia po zakończeniu serii
	// Czas: 5 * 50ms (zapisy) + 500ms (debounce) + margines
	waitForEvent(t, svc.Events, filePath, 2*time.Second)

	// Upewniamy się, że nie ma duplikatów
	assertNoEvent(t, svc.Events, 200*time.Millisecond)
}

// 2. TEST: RECURSIVE WATCH & IGNORE LOGIC
func TestWatcher_Integrity_Recursive(t *testing.T) {

	_, queries := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	root := t.TempDir()

	// Tworzymy strukturę folderów PRZED startem serwisu
	subDir := filepath.Join(root, "models")
	os.Mkdir(subDir, 0755)

	// Dodajemy root do bazy
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queries.CreateScanFolder(ctx, root)

	// Tworzymy Config ręcznie dla tego testu
	cfg := config.NewScannerConfig()
	cfg.SetAllowedExtensions([]string{".png"})

	// Startujemy serwis (Wstrzykujemy config)
	svc, _ := NewService(queries, logger, cfg)

	svc.Startup(ctx)
	defer svc.Shutdown()

	// Dajemy czas na initFolders (WalkDir)
	time.Sleep(200 * time.Millisecond)

	// Tworzymy plik w podkatalogu, który powinien być już obserwowany
	filePath := filepath.Join(subDir, "hero.png")
	createDummyFile(t, filePath)

	waitForEvent(t, svc.Events, filePath, 2*time.Second)
}

// 3. TEST: UNWATCH
// Sprawdzamy czy Unwatch faktycznie przestaje słuchać.
func TestWatcher_Integrity_Unwatch(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	svc.Startup(ctx)
	svc.Watch(root)
	time.Sleep(100 * time.Millisecond)

	// Upewniamy się, że działa
	file1 := filepath.Join(root, "should_detect.png")
	createDummyFile(t, file1)
	waitForEvent(t, svc.Events, file1, 1*time.Second)

	// Wyłączamy obserwację
	svc.Unwatch(root)
	time.Sleep(100 * time.Millisecond) // Czas na propagację w fsnotify

	// To nie powinno zostać wykryte
	file2 := filepath.Join(root, "should_ignore.png")
	createDummyFile(t, file2)

	assertNoEvent(t, svc.Events, 500*time.Millisecond)
}

// 4. TEST: EXTENSION FILTERING
// Sprawdzamy czy ignoruje pliki spoza allowedExtensions
func TestWatcher_Filter_Extensions(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	// Konfigurujemy tylko .png (zrobione w setupWatcherTest, ale dla pewności powtarzamy poprawną metodą)
	svc.config.SetAllowedExtensions([]string{".png"})

	svc.Startup(ctx)
	time.Sleep(100 * time.Millisecond)

	// 1. Plik .txt (powinien być zignorowany)
	txtFile := filepath.Join(root, "notes.txt")
	createDummyFile(t, txtFile)

	// 2. Plik .png (powinien być wykryty)
	pngFile := filepath.Join(root, "image.png")
	createDummyFile(t, pngFile)

	// Oczekujemy tylko PNG
	waitForEvent(t, svc.Events, pngFile, 1*time.Second)

	// Upewniamy się, że TXT nie wpadł "przypadkiem" wcześniej
	assertNoEvent(t, svc.Events, 100*time.Millisecond)
}
