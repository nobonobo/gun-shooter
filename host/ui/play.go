package ui

import (
	"fmt"
	"log"
	"math/rand"
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

	textFont     *ui.Font
	screenWidth  int
	screenHeight int

	cnt int

	particles      []particle
	lastUpdateTime time.Time

	globalState GlobalState
}

type particle struct {
	x, y   float32
	vx, vy float32
	life   float32 // 1.0 down to 0.0
}

var _ ui.ElementKeyboardHandler = (*playScreenComponent)(nil)

func (c *playScreenComponent) OnCreate() {
	c.debugVisible = false

	c.globalState = co.TypedValue[GlobalState](c.Scope())
	c.engine = c.globalState.Engine
	c.resourceSet = c.globalState.ResourceSet

	componentData := co.GetData[PlayScreenData](c.Properties())
	c.app = componentData.App

	c.textFont = co.OpenFont(c.Scope(), "ui:///roboto-regular.ttf")
	c.screenWidth = 1280
	c.screenHeight = 840
	c.lastUpdateTime = time.Now()

	c.createScene()
	c.engine.SetActiveScene(c.scene)
	c.engine.ResetDeltaTime()

	Fullscreen(true)
}

var _ ui.ElementRenderHandler = (*playScreenComponent)(nil)

func (c *playScreenComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	// 画面サイズを要素の現在のサイズに同期
	c.screenWidth = element.Bounds().Width
	c.screenHeight = element.Bounds().Height

	now := time.Now()
	dt := float32(now.Sub(c.lastUpdateTime).Seconds())
	c.lastUpdateTime = now

	// 100フレームごとのデバッグログ
	c.cnt++
	logDebug := c.cnt%100 == 0

	for id, active := range c.globalState.Actives {
		if time.Since(active.Time) > 5*time.Second {
			continue
		}
		if logDebug {
			log.Println("active:", id, active.Info.Name, active.Info.X, active.Info.Y, active.Info.Fire)
		}

		// Fire == true の場合にパーティクルを生成
		if active.Info.Fire {
			x := float32(active.Info.X * float64(c.screenWidth))
			y := float32(active.Info.Y * float64(c.screenHeight))
			for i := 0; i < 5; i++ {
				c.particles = append(c.particles, particle{
					x:    x,
					y:    y,
					vx:   (rand.Float32() - 0.5) * 500,
					vy:   (rand.Float32() - 0.5) * 500,
					life: 1.0,
				})
			}
		}
	}

	// パーティクルの更新
	for i := 0; i < len(c.particles); {
		p := &c.particles[i]
		p.x += p.vx * dt
		p.y += p.vy * dt
		p.life -= dt * 3.0 // 約0.33秒で消える
		if p.life <= 0 {
			c.particles[i] = c.particles[len(c.particles)-1]
			c.particles = c.particles[:len(c.particles)-1]
		} else {
			i++
		}
	}

	// パーティクルの描画
	for _, p := range c.particles {
		color := ui.RGBA(255, 128, 0, uint8(p.life*255)) // オレンジ色からフェードアウト
		canvas.FillTextLine([]rune("*"), sprec.Vec2{X: p.x, Y: p.y}, ui.Typography{
			Font:  c.textFont,
			Size:  24.0,
			Color: color,
		})
	}

	c.Invalidate()
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

		// Player Markers
		for id, active := range c.globalState.Actives {
			if time.Since(active.Time) > 5*time.Second {
				continue
			}
			x := int(active.Info.X * float64(c.screenWidth))
			y := int(active.Info.Y * float64(c.screenHeight))

			co.WithChild("player-"+id, co.New(std.Element, func() {
				co.WithLayoutData(layout.Data{
					HorizontalCenter: opt.V(x - c.screenWidth/2),
					Top:              opt.V(y - 5),
				})
				co.WithData(std.ElementData{
					Layout: layout.Vertical(layout.VerticalSettings{
						ContentAlignment: layout.HorizontalAlignmentCenter,
					}),
				})

				co.WithChild("dot", co.New(std.Container, func() {
					color := ui.Green()
					if active.Info.Fire {
						color = ui.Red()
					}
					co.WithLayoutData(layout.Data{
						Width:  opt.V(10),
						Height: opt.V(10),
					})
					co.WithData(std.ContainerData{
						BackgroundColor: opt.V(color),
					})
				}))

				co.WithChild("label", co.New(std.Label, func() {
					co.WithData(std.LabelData{
						Font:      c.textFont,
						FontSize:  opt.V(float32(16)),
						FontColor: opt.V(ui.White()),
						Text:      active.Info.Name,
					})
				}))
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
