package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	store, err := openStore()
	if err != nil {
		log.Fatal(err)
	}
	bridge := wailsbridge.NewBridge(store)

	err = wails.Run(&options.App{
		Title:     "NetSight",
		Width:     1400,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  bridge.SetContext,
		OnShutdown: func(ctx context.Context) { bridge.Shutdown(); _ = store.Close() },
		Bind: []interface{}{
			bridge,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func openStore() (storage.Store, error) {
	if os.Getenv("NETSIGHT_MEMORY_STORE") == "1" {
		log.Println("WARNING: using in-memory store because NETSIGHT_MEMORY_STORE=1; data will not persist")
		return storage.NewMemoryStore(), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir, err = os.UserHomeDir()
	}
	if err != nil {
		return nil, fmt.Errorf("cannot determine application data directory: %w (set NETSIGHT_MEMORY_STORE=1 for explicit in-memory fallback)", err)
	}
	dataDir := filepath.Join(configDir, "NetSight")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create application data directory: %w", err)
	}
	store, err := storage.NewSQLiteStore(filepath.Join(dataDir, "netsight.db"))
	if err != nil {
		return nil, fmt.Errorf("persistent SQLite store unavailable: %w (set NETSIGHT_MEMORY_STORE=1 for explicit in-memory fallback)", err)
	}
	return store, nil
}
