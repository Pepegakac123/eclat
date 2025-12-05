package main

import (
	"embed"
	"log"

	"database/sql"
	"eclat/internal/app"

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

	db, err := sql.Open("sqlite", "assets.db?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "sql/schema"); err != nil {
		log.Fatal(err)
	}

	myApp := app.NewApp(db)
	err = wails.Run(&options.App{
		Title:  "Eclat",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        myApp.OnStartup,
		Bind: []interface{}{
			myApp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
