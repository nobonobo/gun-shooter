//go:build js

package main

import (
	"fmt"

	jsapp "github.com/mokiat/lacking-js/app"
	jsgame "github.com/mokiat/lacking-js/game"
	jsui "github.com/mokiat/lacking-js/ui"
	"github.com/mokiat/lacking/storage/chunked"
)

func runApplication() error {
	storage, err := chunked.NewWebStorage(".")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	controller := createController(storage, jsgame.NewShaderCollection(), jsgame.NewShaderBuilder(), jsui.NewShaderCollection())

	cfg := jsapp.NewConfig("screen")
	cfg.AddGLExtension("EXT_color_buffer_float")
	cfg.SetFullscreen(false)
	cfg.SetAudioEnabled(true)
	return jsapp.Run(cfg, controller)
}
