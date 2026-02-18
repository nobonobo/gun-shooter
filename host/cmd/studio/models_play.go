package main

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/game/asset/mdl"
)

var _ = func() any {
	skyImage := dsl.CubeImageFromEquirectangular(
		dsl.OpenImage("resources/raw/images/skybox.exr"),
	)
	skyImageSmall := dsl.ResizedCubeImage(skyImage, dsl.Const(128))

	smallerSkyImage := dsl.ResizedCubeImage(skyImage, dsl.Const(512))
	skyTexture := dsl.CreateCubeTexture(smallerSkyImage)

	reflectionCubeImages := dsl.ReflectionCubeImages(skyImageSmall, dsl.SetSampleCount(dsl.Const(120)))
	reflectionTexture := dsl.CreateCubeMipmapTexture(reflectionCubeImages,
		dsl.SetMipmapping(dsl.Const(true)),
	)

	refractionCubeImage := dsl.IrradianceCubeImage(skyImageSmall, dsl.SetSampleCount(dsl.Const(50)))
	refractionTexture := dsl.CreateCubeTexture(refractionCubeImage)

	skyMaterial := dsl.CreateTextureSkyMaterial(
		dsl.CreateSampler(skyTexture,
			dsl.SetWrapMode(dsl.Const(mdl.WrapModeClamp)),
			dsl.SetFilterMode(dsl.Const(mdl.FilterModeLinear)),
			dsl.SetMipmapping(dsl.Const(false)),
		),
	)

	sky := dsl.CreateSky(skyMaterial)

	ambientLight := dsl.CreateAmbientLight(
		dsl.SetReflectionTexture(reflectionTexture),
		dsl.SetRefractionTexture(refractionTexture),
	)

	directionalLight := dsl.CreateDirectionalLight(
		dsl.SetEmitColor(dsl.RGB(1.1, 1.0, 1.3)),
		dsl.SetCastShadow(dsl.Const(true)),
	)

	return dsl.Save("play-screen.dat", dsl.CreateModel(
		dsl.AddNode(dsl.CreateNode("Sky",
			dsl.AddAttachment(sky),
		)),
		dsl.AddNode(dsl.CreateNode("AmbientLight",
			dsl.AddAttachment(ambientLight),
		)),
		dsl.AddNode(dsl.CreateNode("DirectionalLight",
			dsl.AddAttachment(directionalLight),
			dsl.SetRotation(dsl.Const(dprec.QuatProd(
				dprec.RotationQuat(dprec.Degrees(-30), dprec.BasisYVec3()),
				dprec.RotationQuat(dprec.Degrees(-45), dprec.BasisXVec3()),
			))),
		)),
	))
}()

var _ = dsl.Save("board.dat",
	dsl.OpenGLTFModel("resources/raw/models/board.glb"),
)

var _ = dsl.Save("ball.dat",
	dsl.OpenGLTFModel("resources/raw/models/ball.glb"),
)
