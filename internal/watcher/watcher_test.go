package watcher

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// 2. TEST: FILTROWANIE
// Tworzymy pliki, które powinny być ignorowane.
func TestWatcher_Integrity_Filtering(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	svc.Startup(ctx)
	time.Sleep(100 * time.Millisecond)

	// A. Plik tekstowy (nie ma go w AllowedExtensions)
	txtPath := filepath.Join(root, "notes.txt")
	createDummyFile(t, txtPath)

	// B. Plik tymczasowy / ukryty (z kropką)
	hiddenPath := filepath.Join(root, ".temp.png")
	createDummyFile(t, hiddenPath)

	// Oczekujemy CISZY na kanale
	assertNoEvent(t, svc.Events, 1*time.Second)

	// C. Plik poprawny (kontrola, czy watcher w ogóle działa)
	goodPath := filepath.Join(root, "valid.png")
	createDummyFile(t, goodPath)
	waitForEvent(t, svc.Events, goodPath, 1*time.Second)
}

// 3. TEST: REKURENCJA & DYNAMICZNE DODAWANIE
// Tworzymy folder, a w nim plik. Watcher musi to wyłapać w locie.
func TestWatcher_Integrity_RecursiveCreate(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	svc.Startup(ctx)
	time.Sleep(100 * time.Millisecond)

	// 1. Tworzymy podkatalog
	subDir := filepath.Join(root, "textures")
	err := os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	// Ważne: Dajemy watcherowi chwilę na zarejestrowanie nowego folderu
	time.Sleep(200 * time.Millisecond)

	// 2. Tworzymy plik w tym nowym podkatalogu
	filePath := filepath.Join(subDir, "wood.png")
	createDummyFile(t, filePath)

	// 3. Sprawdzamy czy Watcher wysłał zdarzenie
	waitForEvent(t, svc.Events, filePath, 2*time.Second)
}

// 4. TEST: INIT FOLDERS
// Sprawdzamy, czy przy starcie watcher widzi istniejące pliki/foldery z bazy.
func TestWatcher_Integrity_InitFolders(t *testing.T) {

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

	// Startujemy serwis
	svc, _ := NewService(queries, logger)
	svc.config.AllowedExtensions = []string{".png"}
	svc.Startup(ctx)
	defer svc.Shutdown()

	// Dajemy czas na initFolders (WalkDir)
	time.Sleep(200 * time.Millisecond)

	// Tworzymy plik w podkatalogu, który powinien być już obserwowany
	filePath := filepath.Join(subDir, "hero.png")
	createDummyFile(t, filePath)

	waitForEvent(t, svc.Events, filePath, 2*time.Second)
}

// 5. TEST: UNWATCH
// Sprawdzamy czy Unwatch faktycznie przestaje słuchać.
func TestWatcher_Integrity_Unwatch(t *testing.T) {
	svc, _, root, ctx, cancel := setupWatcherTest(t)
	defer cancel()
	defer svc.Shutdown()

	svc.Startup(ctx)
	svc.Watch(root)
	time.Sleep(100 * time.Millisecond)

	// 1. Sprawdzamy czy działa
	file1 := filepath.Join(root, "test1.png")
	createDummyFile(t, file1)
	waitForEvent(t, svc.Events, file1, 1*time.Second)

	// 2. Robimy UNWATCH
	svc.Unwatch(root)
	time.Sleep(100 * time.Millisecond)

	// 3. Tworzymy kolejny plik - Watcher powinien go Olać
	file2 := filepath.Join(root, "test2.png")
	createDummyFile(t, file2)

	assertNoEvent(t, svc.Events, 1*time.Second)
}
