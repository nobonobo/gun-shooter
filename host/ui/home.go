package ui

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"

	"github.com/mokiat/lacking/util/async"

	"github.com/nobonobo/gun-shooter/host/ui/widget"
)

func LoadHomeData(engine *game.Engine, resourceSet *game.ResourceSet) async.Promise[*HomeData] {
	var data HomeData
	return async.InjectionPromise(async.JoinOperations(
		resourceSet.FetchResource("home-screen.dat", &data.Scene),
	), &data)
}

type HomeData struct {
	Scene *game.ModelTemplate
}

var HomeScreen = co.Define[*homeScreenComponent]()

type HomeScreenData struct {
	App *applicationComponent
}

type homeScreenComponent struct {
	co.BaseComponent

	app *applicationComponent

	engine      *game.Engine
	resourceSet *game.ResourceSet

	sceneData *HomeData
	scene     *game.Scene
}

func (c *homeScreenComponent) OnCreate() {
	globalState := co.TypedValue[GlobalState](c.Scope())
	c.engine = globalState.Engine
	c.resourceSet = globalState.ResourceSet

	componentData := co.GetData[HomeScreenData](c.Properties())
	c.app = componentData.App

	// In a real app we might want to preserve this state
	// but for now we'll just create it fresh
	c.createScene()

	c.engine.SetActiveScene(c.scene)
	c.engine.ResetDeltaTime()
}

func (c *homeScreenComponent) OnDelete() {
	c.engine.SetActiveScene(nil)
}

func (c *homeScreenComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithData(std.ElementData{
			Layout: layout.Anchor(),
		})

		co.WithChild("pane", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Bottom: opt.V(0),
				Left:   opt.V(0),
				Width:  opt.V(320),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.RGBA(0, 0, 0, 192)),
				Layout:          layout.Anchor(),
			})

			co.WithChild("holder", co.New(std.Element, func() {
				co.WithLayoutData(layout.Data{
					Left:           opt.V(75),
					VerticalCenter: opt.V(0),
				})
				co.WithData(std.ElementData{
					Layout: layout.Vertical(layout.VerticalSettings{
						ContentAlignment: layout.HorizontalAlignmentLeft,
						ContentSpacing:   15,
					}),
				})

				co.WithChild("play-button", co.New(widget.Button, func() {
					co.WithData(widget.ButtonData{
						Text: "Play",
					})
					co.WithCallbackData(widget.ButtonCallbackData{
						OnClick: c.onPlayClicked,
					})
				}))

				co.WithChild("licenses-button", co.New(widget.Button, func() {
					co.WithData(widget.ButtonData{
						Text: "Licenses",
					})
					co.WithCallbackData(widget.ButtonCallbackData{
						OnClick: c.onLicensesClicked,
					})
				}))

				co.WithChild("exit-button", co.New(widget.Button, func() {
					co.WithData(widget.ButtonData{
						Text: "Exit",
					})
					co.WithCallbackData(widget.ButtonCallbackData{
						OnClick: c.onExitClicked,
					})
				}))
			}))
		}))
	})
}

func (c *homeScreenComponent) createScene() {
	c.sceneData = homeSceneData // retrieve from global storage or similar if needed in future, currently hacky sharing

	c.scene = c.engine.CreateScene(game.SceneInfo{
		IncludePhysics: opt.V(false),
		IncludeECS:     opt.V(false),
	})

	sceneModel := c.scene.InstantiateModel(game.ModelInfo{
		Template:  c.sceneData.Scene,
		Name:      opt.V("Scene"),
		IsDynamic: false,
	})

	camera := c.createCamera(c.scene.Graphics())
	c.scene.Graphics().SetActiveCamera(camera)

	if cameraNode := sceneModel.FindNode("Camera"); !cameraNode.IsNil() {
		c.scene.CameraBindingSet().Bind(cameraNode, camera)
	}

	const animationName = "CameraRotation"
	if recording := sceneModel.FindRecording(animationName); recording != nil {
		playback := recording.Playback(true)
		player := sceneModel.BindAnimation(playback)
		c.scene.PlayAnimation(player)
	}
}

func (c *homeScreenComponent) createCamera(scene *graphics.Scene) *graphics.Camera {
	result := scene.CreateCamera()
	result.SetFoVMode(graphics.FoVModeHorizontalPlus)
	result.SetFoV(sprec.Degrees(30))
	result.SetAutoExposure(false)
	result.SetExposure(1.0)
	result.SetAutoFocus(false)
	result.SetAutoExposureSpeed(0.1)
	result.SetCascadeDistances([]float32{32.0})
	return result
}

func (c *homeScreenComponent) onPlayClicked() {
	promise := NewLoadingPromise(
		co.Window(c.Scope()),
		LoadPlayData(c.engine, c.resourceSet),
		func(d *PlayData) {
			playSceneData = d
		},
		func(err error) {
			loadingError = err
		},
	)
	loadingState = LoadingState{
		Promise:         promise,
		SuccessViewName: ViewNamePlay,
		ErrorViewName:   ViewNameError,
	}
	c.app.SetActiveView(ViewNameLoading)
}

func (c *homeScreenComponent) onLicensesClicked() {
	c.app.SetActiveView(ViewNameLicenses)
}

func (c *homeScreenComponent) onExitClicked() {
	co.Window(c.Scope()).Close()
}

// Temporary global storage for data across views, replacing strict models
var homeSceneData *HomeData
