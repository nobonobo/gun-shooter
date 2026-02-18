//go:build !js

package main

import (
	"fmt"

	nativeapp "github.com/mokiat/lacking-native/app"
	nativegame "github.com/mokiat/lacking-native/game"
	nativeui "github.com/mokiat/lacking-native/ui"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/resource"

	"github.com/nobonobo/gun-shooter/host/resources"
)

func runApplication() error {
	storage, err := chunked.NewFileStorage("./assets")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	controller := createController(storage, nativegame.NewShaderCollection(), nativegame.NewShaderBuilder(), nativeui.NewShaderCollection())

	cfg := nativeapp.NewConfig("Game", 1280, 800)
	cfg.SetFullscreen(false)
	cfg.SetMaximized(false)
	cfg.SetMinSize(1024, 576)
	cfg.SetVSync(true)
	cfg.SetIcon("ui/images/icon.png")
	cfg.SetLocator(ui.WrappedLocator(resource.NewFSLocator(resources.UI)))
	cfg.SetAudioEnabled(false)
	return nativeapp.Run(cfg, controller)
}
