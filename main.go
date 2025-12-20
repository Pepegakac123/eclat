package main

import (
	"embed"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"database/sql"
	"eclat/internal/app"
	"eclat/internal/database"
	"eclat/internal/services"

	"github.com/pressly/goose/v3"
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
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Cannot get cache dir:", err)
	}
	// Główny folder: .../AppData/Local/eclat (Windows) lub ~/Library/Caches/eclat (macOS)
	appCachePath := filepath.Join(userCacheDir, "eclat")

	dbFolder := filepath.Join(appCachePath, "db")
	thumbsFolder := filepath.Join(appCachePath, "thumbnails")

	dirsToCreate := []string{appCachePath, dbFolder, thumbsFolder}
	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Cannot create directory %s: %v", dir, err)
		}
	}
	dbPath := filepath.Join(dbFolder, "assets.db")

	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode=WAL")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("Failed to set dialect:", err)
	}
	if err := goose.Up(db, "sql/schema"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	queries := database.New(db)

	thumbGen := services.NewThumbnailGenerator(thumbsFolder, slog.Default())
	scannerService := services.NewScanner(db, queries, thumbGen)
	settingsService := services.NewSettingsService(queries)

	myApp := app.NewApp(scannerService, settingsService)

	err = wails.Run(&options.App{
		Title:            "Eclat",
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DragAndDrop: &options.DragAndDrop{EnableFileDrop: true},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "e7b8a9-eclat",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				myApp.RestoreWindow()
			},
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        myApp.OnStartup,
		Bind: []interface{}{
			myApp,
			scannerService,
			settingsService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
