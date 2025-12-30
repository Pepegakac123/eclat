package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"eclat/internal/bootstrap"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	_ "modernc.org/sqlite"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

func main() {
	// Initialize application dependencies (DB, Config, Services)
	deps, err := bootstrap.Initialize(embedMigrations)
	if err != nil {
		log.Fatalf("Fatal Error during initialization: %v", err)
	}
	// Ensure DB is closed when main exits
	defer deps.DB.Close()

	// Run Wails application
	err = wails.Run(&options.App{
		Title:            "Eclat",
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if after, ok := strings.CutPrefix(r.URL.Path, "/thumbnails/"); ok {
					filename := after
					fullPath := filepath.Join(deps.ThumbnailsDir, filename)
					
					// Check if file exists
					if _, err := os.Stat(fullPath); os.IsNotExist(err) {
						deps.Logger.Error("Thumbnail file does not exist", "url", r.URL.Path, "fullPath", fullPath)
						http.NotFound(w, r)
						return
					}

					deps.Logger.Debug("Serving thumbnail", "url", r.URL.Path, "fullPath", fullPath)
					http.ServeFile(w, r, fullPath)
					return
				}
				deps.Logger.Debug("Asset not found in custom handler", "url", r.URL.Path)
				http.NotFound(w, r)
			}),
		},
		DragAndDrop: &options.DragAndDrop{EnableFileDrop: true},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "e7b8a9-eclat",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				deps.App.RestoreWindow()
			},
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        deps.App.OnStartup,
		OnShutdown:       deps.App.Shutdown,
		Bind: []interface{}{
			deps.App,
			deps.AssetService,
			deps.MaterialSetService,
			deps.ScannerService,
			deps.SettingsService,
			deps.WatcherService,
		},
	})

	if err != nil {
		deps.Logger.Error("Fatal Error", "error", err)
	}
}
