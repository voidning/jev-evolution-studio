package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func frontendFS() fs.FS { sub, _ := fs.Sub(assets, "frontend/dist"); return sub }
func main() {
	headless := flag.Bool("headless", false, "Run the same editor backend with a browser console")
	project := flag.String("project", "", "Explicit initial project path")
	flag.Parse()
	a := NewApp()
	a.headless = *headless
	if *headless {
		a.startup(context.Background())
		defer a.shutdown(context.Background())
		if *project != "" {
			if _, e := a.OpenProject(*project); e != nil {
				fmt.Fprintln(os.Stderr, e)
				return
			}
			if _, e := a.StartPreview(); e != nil {
				fmt.Fprintln(os.Stderr, e)
				return
			}
		}
		fmt.Println("Console:", a.baseURL+"/console/?token="+a.token)
		fmt.Println("Preview:", a.baseURL+"/?token="+a.token)
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		return
	}
	err := wails.Run(&options.App{Title: "Jev — 网页编辑器", Width: 920, Height: 900, MinWidth: 390, MinHeight: 600, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 245, G: 247, B: 248, A: 1}, OnStartup: a.startup, OnShutdown: a.shutdown, Bind: []interface{}{a}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
