package main

import (
	"embed"
	"log"
	"net/http"
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
				if strings.HasPrefix(r.URL.Path, "/thumbnails/") {
					filename := strings.TrimPrefix(r.URL.Path, "/thumbnails/")
					fullPath := filepath.Join(deps.ThumbnailsDir, filename)
					http.ServeFile(w, r, fullPath)
					return
				}
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
