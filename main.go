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
	// 1. DATABASE SETUP
	db, err := sql.Open("sqlite", "assets.db?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	// 2. MIGRATIONS
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("Failed to set dialect:", err)
	}
	if err := goose.Up(db, "sql/schema"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 3. SERVICE SETUP (Dependency Injection)
	queries := database.New(db)

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Cannot get cache dir:", err)
	}
	appCachePath := filepath.Join(userCacheDir, "eclat", "thumbnails")
	if err := os.MkdirAll(appCachePath, 0755); err != nil {
		log.Fatal("Cannot create cache dir:", err)
	}

	thumbGen := services.NewThumbnailGenerator(appCachePath, slog.Default())
	scannerService := services.NewScanner(queries, thumbGen)
	settingsService := services.NewSettingsService(queries)

	// 4. APP SETUP
	myApp := app.NewApp(scannerService, settingsService)

	// 5. WAILS RUN
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

		OnStartup: myApp.OnStartup,

		// BINDING - To co widzi Frontend
		Bind: []interface{}{
			myApp,          // Metody App (jeśli jakieś publiczne będą)
			scannerService, // Metody Scannera (StartScan, GetFolders itp.)
			settingsService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
