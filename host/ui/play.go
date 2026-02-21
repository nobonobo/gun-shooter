package ui

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/debug/metric/metricui"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/acceleration"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
	"github.com/mokiat/lacking/util/async"
	"github.com/mokiat/lacking/util/shape3d"

	"github.com/nobonobo/gun-shooter/host/resources"
)

func FetchSound(audioAPI audio.API, engine *game.Engine, name string, target *audio.Media) async.Operation {
	return async.NewFuncOperation(func() error {
		file, err := resources.UI.Open(name)
		if err != nil {
			log.Printf("ERROR: failed to open sound file %s: %v", name, err)
			return err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			log.Printf("ERROR: failed to read sound data from %s: %v", name, err)
			return err
		}
		log.Printf("pop data: %d bytes", len(data))
		err = engine.ScheduleIO(func() error {
			*target = audioAPI.CreateMedia(audio.MediaInfo{
				Data:     data,
				DataType: audio.MediaDataTypeMP3,
			})
			log.Printf("DEBUG: Created audio media for %s", name)
			return nil
		}).Wait()
		if err != nil {
			log.Printf("ERROR: failed to create audio media for %s: %v", name, err)
			return err
		}

		return nil
	})
}

func LoadPlayData(audioAPI audio.API, engine *game.Engine, resourceSet *game.ResourceSet) async.Promise[*PlayData] {
	var data PlayData
	return async.InjectionPromise(async.JoinOperations(
		resourceSet.FetchResource("play-screen.dat", &data.Scene),
		resourceSet.FetchResource("board.dat", &data.Board),
		resourceSet.FetchResource("ball.dat", &data.Ball),
		FetchSound(audioAPI, engine, "ui/sounds/pop.mp3", &data.Pop),
	), &data)
}

type PlayData struct {
	Scene *game.ModelTemplate
	Board *game.ModelTemplate
	Ball  *game.ModelTemplate
	Pop   audio.Media
}

var PlayScreen = co.Define[*playScreenComponent]()

type PlayScreenData struct {
	App *applicationComponent
}

type PlayMode int

const (
	PlayModeCalibration PlayMode = iota
	PlayModeCountdown
	PlayModePlaying
	PlayModeGameOver
)

type playScreenComponent struct {
	co.BaseComponent

	app *applicationComponent

	debugVisible bool

	audioAPI    audio.API
	engine      *game.Engine
	resourceSet *game.ResourceSet

	sceneData *PlayData
	scene     *game.Scene
	popSound  audio.Media

	textFont     *ui.Font
	screenWidth  int
	screenHeight int

	cnt int

	particles      []particle
	lastUpdateTime time.Time

	globalState GlobalState

	// Game State
	mode     PlayMode
	modeTime time.Duration
	scores   map[string]int
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
	c.audioAPI = c.globalState.AudioAPI
	if c.audioAPI == nil {
		log.Printf("ERROR: audioAPI is nil in playScreenComponent")
	}
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

	c.mode = PlayModeCalibration
	c.scores = make(map[string]int)

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

	// タイマー更新
	if c.modeTime > 0 {
		c.modeTime -= time.Duration(dt * float32(time.Second))
		if c.modeTime <= 0 {
			c.modeTime = 0
			// モード遷移
			switch c.mode {
			case PlayModeCountdown:
				c.mode = PlayModePlaying
				c.modeTime = 60 * time.Second // ゲーム時間は60秒
			case PlayModePlaying:
				c.mode = PlayModeGameOver
			}
		}
	}

	for id, active := range c.globalState.Actives {
		if time.Since(active.Time) > 5*time.Second {
			continue
		}
		if logDebug {
			log.Println("active:", id, active.Info.Name, active.Info.X, active.Info.Y, active.Info.Fire)
		}

		// Fire == true の場合にパーティクルを生成
		if active.Info.Fire {
			active.Info.Fire = false
			x := float32(active.Info.X * float64(c.screenWidth))
			y := float32(active.Info.Y * float64(c.screenHeight))
			if x < 0 || y < 0 || x > float32(c.screenWidth) || y > float32(c.screenHeight) {
				continue
			}

			// プレイ中のみスコア加算
			if c.mode == PlayModePlaying {
				c.scores[id]++
			}

			c.audioAPI.Play(c.popSound, audio.PlayInfo{
				Gain: opt.V(1.0),
			})
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
		canvas.Reset()
		canvas.Circle(sprec.Vec2{X: p.x, Y: p.y}, 5)
		canvas.Fill(ui.Fill{
			Color: color,
		})
	}

	invalid := true
	if c.mode == PlayModeCalibration {
		invalid = false // キャリブレーション中は描画更新不要（パーティクルがなければ）
	}
	if len(c.particles) > 0 || c.modeTime > 0 || c.mode == PlayModePlaying {
		invalid = true
	}

	if invalid {
		c.Invalidate()
	}
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

		// Player Markers (Only visible during Calibration, Countdown, and Playing)
		if c.mode != PlayModeGameOver {
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
		}

		// Mode Overlays
		co.WithChild("overlay", co.New(std.Element, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Left:   opt.V(0),
				Right:  opt.V(0),
				Bottom: opt.V(0),
			})
			co.WithData(std.ElementData{
				Layout: layout.Anchor(),
			})

			switch c.mode {
			case PlayModeCalibration:
				co.WithChild("calibration-box", co.New(std.Container, func() {
					co.WithLayoutData(layout.Data{
						HorizontalCenter: opt.V(0),
						VerticalCenter:   opt.V(0),
					})
					co.WithData(std.ContainerData{
						BackgroundColor: opt.V(ui.RGBA(0, 0, 0, 180)),
						Padding:         ui.Spacing{Left: 40, Right: 40, Top: 20, Bottom: 20},
						Layout: layout.Vertical(layout.VerticalSettings{
							ContentAlignment: layout.HorizontalAlignmentCenter,
							ContentSpacing:   20,
						}),
					})

					co.WithChild("text", co.New(std.Label, func() {
						co.WithData(std.LabelData{
							Font:      c.textFont,
							FontSize:  opt.V(float32(32)),
							FontColor: opt.V(ui.White()),
							Text:      "Waiting for Calibration...",
						})
					}))

					co.WithChild("start-btn", co.New(std.Button, func() {
						co.WithData(std.ButtonData{
							Text: "START GAME",
						})
						co.WithCallbackData(std.ButtonCallbackData{
							OnClick: func() {
								c.mode = PlayModeCountdown
								c.modeTime = 3 * time.Second
								c.scores = make(map[string]int)
								c.Invalidate()
							},
						})
					}))
				}))

			case PlayModeCountdown:
				co.WithChild("countdown-text", co.New(std.Label, func() {
					co.WithLayoutData(layout.Data{
						HorizontalCenter: opt.V(0),
						VerticalCenter:   opt.V(0),
					})
					seconds := int(c.modeTime.Seconds()) + 1
					co.WithData(std.LabelData{
						Font:      c.textFont,
						FontSize:  opt.V(float32(128)),
						FontColor: opt.V(ui.Yellow()),
						Text:      fmt.Sprintf("%d", seconds),
					})
				}))

			case PlayModePlaying:
				// HUD
				co.WithChild("hud", co.New(std.Container, func() {
					co.WithLayoutData(layout.Data{
						Top:              opt.V(20),
						HorizontalCenter: opt.V(0),
					})
					co.WithData(std.ContainerData{
						BackgroundColor: opt.V(ui.RGBA(0, 0, 0, 120)),
						Padding:         ui.Spacing{Left: 20, Right: 20, Top: 10, Bottom: 10},
						Layout: layout.Horizontal(layout.HorizontalSettings{
							ContentAlignment: layout.VerticalAlignmentCenter,
							ContentSpacing:   40,
						}),
					})

					co.WithChild("timer", co.New(std.Label, func() {
						co.WithData(std.LabelData{
							Font:      c.textFont,
							FontSize:  opt.V(float32(32)),
							FontColor: opt.V(ui.White()),
							Text:      fmt.Sprintf("TIME: %d", int(c.modeTime.Seconds())),
						})
					}))
				}))

			case PlayModeGameOver:
				co.WithChild("gameover-box", co.New(std.Container, func() {
					co.WithLayoutData(layout.Data{
						HorizontalCenter: opt.V(0),
						VerticalCenter:   opt.V(0),
					})
					co.WithData(std.ContainerData{
						BackgroundColor: opt.V(ui.RGBA(0, 0, 0, 220)),
						Padding:         ui.Spacing{Left: 60, Right: 60, Top: 40, Bottom: 40},
						Layout: layout.Vertical(layout.VerticalSettings{
							ContentAlignment: layout.HorizontalAlignmentCenter,
							ContentSpacing:   20,
						}),
					})

					co.WithChild("title", co.New(std.Label, func() {
						co.WithData(std.LabelData{
							Font:      c.textFont,
							FontSize:  opt.V(float32(48)),
							FontColor: opt.V(ui.Red()),
							Text:      "GAME OVER",
						})
					}))

					// Score List
					co.WithChild("scores", co.New(std.Element, func() {
						co.WithData(std.ElementData{
							Layout: layout.Vertical(layout.VerticalSettings{
								ContentAlignment: layout.HorizontalAlignmentCenter,
								ContentSpacing:   5,
							}),
						})
						for id, score := range c.scores {
							name := "Unknown"
							if active, ok := c.globalState.Actives[id]; ok {
								name = active.Info.Name
							}
							co.WithChild("score-"+id, co.New(std.Label, func() {
								co.WithData(std.LabelData{
									Font:      c.textFont,
									FontSize:  opt.V(float32(24)),
									FontColor: opt.V(ui.White()),
									Text:      fmt.Sprintf("%s: %d", name, score),
								})
							}))
						}
					}))

					co.WithChild("actions", co.New(std.Element, func() {
						co.WithData(std.ElementData{
							Layout: layout.Horizontal(layout.HorizontalSettings{
								ContentSpacing: 20,
							}),
						})
						co.WithChild("restart-btn", co.New(std.Button, func() {
							co.WithData(std.ButtonData{
								Text: "RESTART",
							})
							co.WithCallbackData(std.ButtonCallbackData{
								OnClick: func() {
									c.mode = PlayModeCountdown
									c.modeTime = 3 * time.Second
									c.scores = make(map[string]int)
									c.Invalidate()
								},
							})
						}))
						co.WithChild("exit-btn", co.New(std.Button, func() {
							co.WithData(std.ButtonData{
								Text: "EXIT",
							})
							co.WithCallbackData(std.ButtonCallbackData{
								OnClick: func() {
									co.Window(c.Scope()).Close()
								},
							})
						}))
					}))
				}))
			}
		}))
	})
}

func (c *playScreenComponent) createScene() {
	c.sceneData = playSceneData // retrieve from global storage
	c.popSound = c.sceneData.Pop
	c.audioAPI.Play(c.popSound, audio.PlayInfo{
		Gain: opt.V(1.0),
	})

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
