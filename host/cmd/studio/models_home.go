package main

import (
	"github.com/mokiat/lacking/game/asset/dsl"
)

var _ = func() any {
	sky := dsl.CreateSky(dsl.CreateColorSkyMaterial(
		dsl.RGB(2.0, 1.5, 0.5),
	))

	return dsl.Save("home-screen.dat", dsl.CreateModel(
		dsl.AppendModel(dsl.OpenGLTFModel("resources/raw/models/home.glb")),
		dsl.AddNode(dsl.CreateNode("Sky",
			dsl.AddAttachment(sky),
		)),
	))
}()
