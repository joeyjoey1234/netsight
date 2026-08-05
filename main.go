package main

import (
	"embed"
	"log"
	"os"

	"netsight/internal/storage"
	"netsight/internal/wailsbridge"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if len(os.Args) > 1 {
		runCLI()
		return
	}

	store := storage.NewMemoryStore()
	bridge := wailsbridge.NewBridge(store)

	err := wails.Run(&options.App{
		Title:     "NetSight",
		Width:     1400,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  bridge.SetContext,
		OnShutdown: func(ctx interface{}) { store.Close() },
		Bind: []interface{}{
			bridge,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
