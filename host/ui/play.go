package ui

import (
	"fmt"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/metric/metricui"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/acceleration"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
	"github.com/mokiat/lacking/util/shape3d"

	"github.com/mokiat/lacking/util/async"
)

func LoadPlayData(engine *game.Engine, resourceSet *game.ResourceSet) async.Promise[*PlayData] {
	var data PlayData
	return async.InjectionPromise(async.JoinOperations(
		resourceSet.FetchResource("play-screen.dat", &data.Scene),
		resourceSet.FetchResource("board.dat", &data.Board),
		resourceSet.FetchResource("ball.dat", &data.Ball),
	), &data)
}

type PlayData struct {
	Scene *game.ModelTemplate
	Board *game.ModelTemplate
	Ball  *game.ModelTemplate
}

var PlayScreen = co.Define[*playScreenComponent]()

type PlayScreenData struct {
	App *applicationComponent
}

type playScreenComponent struct {
	co.BaseComponent

	app *applicationComponent

	debugVisible bool

	engine      *game.Engine
	resourceSet *game.ResourceSet

	sceneData *PlayData
	scene     *game.Scene

	globalState GlobalState
}

var _ ui.ElementKeyboardHandler = (*playScreenComponent)(nil)

func (c *playScreenComponent) OnCreate() {
	c.debugVisible = false

	c.globalState = co.TypedValue[GlobalState](c.Scope())
	c.engine = c.globalState.Engine
	c.resourceSet = c.globalState.ResourceSet

	componentData := co.GetData[PlayScreenData](c.Properties())
	c.app = componentData.App

	c.createScene()
	c.engine.SetActiveScene(c.scene)
	c.engine.ResetDeltaTime()
	Fullscreen(true)
}

func (c *playScreenComponent) OnDelete() {
	c.engine.SetActiveScene(nil)
	Fullscreen(false)
}

func (c *playScreenComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Code {

	case ui.KeyCodeEscape:
		co.Window(c.Scope()).Close()
		return true

	case ui.KeyCodeTab:
		if event.Action == ui.KeyboardActionDown {
			c.debugVisible = !c.debugVisible
			c.Invalidate()
		}
		return true

	default:
		return false
	}
}

func (c *playScreenComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithData(std.ElementData{
			Essence:       c,
			CanAutoFocus:  opt.V(true),
			CreateFocused: true,
			Layout:        layout.Anchor(),
		})

		if c.debugVisible {
			co.WithChild("flamegraph", co.New(metricui.FlameGraph, func() {
				co.WithData(metricui.FlameGraphData{
					UpdateInterval: time.Second,
				})
				co.WithLayoutData(layout.Data{
					Top:   opt.V(0),
					Left:  opt.V(0),
					Right: opt.V(0),
				})
			}))
		}

		// Marker Images in Corners
		for i := 0; i < 4; i++ {
			imagePath := fmt.Sprintf("ui/images/pattern-marker_%d.png", i)
			var layoutData layout.Data
			switch i {
			case 0: // Top-Left
				layoutData = layout.Data{
					Top:  opt.V(0),
					Left: opt.V(0),
				}
			case 1: // Top-Right
				layoutData = layout.Data{
					Top:   opt.V(0),
					Right: opt.V(0),
				}
			case 2: // Bottom-Right
				layoutData = layout.Data{
					Bottom: opt.V(0),
					Left:   opt.V(0),
				}
			case 3: // Bottom-Left
				layoutData = layout.Data{
					Bottom: opt.V(0),
					Right:  opt.V(0),
				}
			}
			layoutData.Width = opt.V(200)
			layoutData.Height = opt.V(200)

			co.WithChild(fmt.Sprintf("marker-%d", i), co.New(std.Picture, func() {
				co.WithLayoutData(layoutData)
				co.WithData(std.PictureData{
					BackgroundColor: opt.V(ui.Transparent()),
					Image:           co.OpenImage(c.Scope(), imagePath),
					Mode:            std.ImageModeFit,
				})
			}))
		}
	})
}

func (c *playScreenComponent) createScene() {
	c.sceneData = playSceneData // retrieve from global storage

	c.scene = c.engine.CreateScene(game.SceneInfo{
		IncludeECS: opt.V(false),
	})

	c.scene.InstantiateModel(game.ModelInfo{
		Template:  c.sceneData.Scene,
		Name:      opt.V("Scene"),
		IsDynamic: false,
	})

	boardModel := c.scene.InstantiateModel(game.ModelInfo{
		Template:  c.sceneData.Board,
		Name:      opt.V("Board"),
		IsDynamic: false,
	})

	camera := c.createCamera(c.scene.Graphics())
	c.scene.Graphics().SetActiveCamera(camera)

	if cameraNode := boardModel.FindNode("Camera"); !cameraNode.IsNil() {
		c.scene.CameraBindingSet().Bind(cameraNode, camera)
	}

	ballModel := c.scene.InstantiateModel(game.ModelInfo{
		Template:  c.sceneData.Ball,
		Name:      opt.V("Ball"),
		Position:  opt.V(dprec.NewVec3(-1.0, 3.0, 2.0)),
		IsDynamic: true,
	})
	ballModelNode := c.scene.Hierarchy().Wrap(ballModel.Root())

	physicsScene := c.scene.Physics()
	ballBodyDef := physicsScene.Engine().CreateBodyDefinition(physics.BodyDefinitionInfo{
		Mass:                   1.0,
		MomentOfInertia:        physics.SolidSphereMomentOfInertia(1.0, 1.0),
		FrictionCoefficient:    0.5,
		RestitutionCoefficient: 0.5,
		DragFactor:             0.1,
		AngularDragFactor:      0.1,
		CollisionGroup:         1,
		CollisionSpheres: []shape3d.Sphere{
			shape3d.NewSphere(dprec.ZeroVec3(), 1.0),
		},
	})
	ballBody := physicsScene.CreateBody(physics.BodyInfo{
		Name:       "Ball",
		Definition: ballBodyDef,
		Position:   ballModelNode.Position(),
		Rotation:   ballModelNode.Rotation(),
	})
	ballBody.SetVelocity(dprec.NewVec3(0.0, 0.0, 3.0))
	c.scene.BodyBindingSet().Bind(ballModelNode.ID(), ballBody)

	physicsScene.CreateGlobalAccelerator(acceleration.NewGravityDirection())
}

func (c *playScreenComponent) createCamera(scene *graphics.Scene) *graphics.Camera {
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

// Temporary global storage for data across views
var playSceneData *PlayData
