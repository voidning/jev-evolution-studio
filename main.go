package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title: "Forge — Evolution Director", Width: 560, Height: 900, MinWidth: 480, MinHeight: 720,
		AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 9, G: 9, B: 12, A: 1},
		OnStartup: app.startup, OnShutdown: app.shutdown, Bind: []interface{}{app},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
