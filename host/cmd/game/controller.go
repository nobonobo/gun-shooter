package main

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/resource"

	"github.com/nobonobo/gun-shooter/host/resources"
	gameui "github.com/nobonobo/gun-shooter/host/ui"
)

func createController(storage chunked.Storage, gameShaders graphics.ShaderCollection, gameBuilder graphics.ShaderBuilder, uiShaders ui.ShaderCollection) app.Controller {
	locator := ui.WrappedLocator(resource.NewFSLocator(resources.UI))

	gameController := game.NewController(storage, gameShaders, gameBuilder)
	uiController := ui.NewController(locator, uiShaders, func(w *ui.Window) {
		gameui.BootstrapApplication(w, gameController)
	})

	return app.NewLayeredController(gameController, uiController)
}
