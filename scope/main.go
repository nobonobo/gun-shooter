package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"syscall/js"
	"time"

	"github.com/google/uuid"
	"github.com/nobonobo/gun-shooter/schema"
	"github.com/nobonobo/rtcconnect/node"
	"github.com/pion/webrtc/v4"
)

type Application struct {
	scene        js.Value
	camera       js.Value
	renderer     js.Value
	arToolkitCtx js.Value
	arToolkitSrc js.Value
	markers      []js.Value
	patternUrls  []string
	uid          string
	name         string
	dest         string
	node         *node.Node
	ctx          context.Context
	cancel       context.CancelFunc
	cnt          int
	fire         bool
	OnUpdate     func([4]Marker)
}

func NewApplication() *Application {
	uid, _ := uuid.NewV6()
	u, _ := url.Parse(location.Get("href").String())
	name := u.Query().Get("name")
	if name == "" {
		name = "NoName"
	}
	dest := u.Query().Get("dest")
	n := node.New(uid.String())
	app := &Application{
		patternUrls: []string{
			"marker/pattern-marker_0.patt",
			"marker/pattern-marker_1.patt",
			"marker/pattern-marker_3.patt",
			"marker/pattern-marker_2.patt",
		},
		uid:      uid.String(),
		name:     name,
		dest:     dest,
		node:     n,
		ctx:      context.Background(),
		cancel:   func() {},
		OnUpdate: func(markers [4]Marker) {},
	}
	return app
}

func (app *Application) Publish(data []byte, force bool) error {
	app.cnt++
	if app.cnt%100 == 0 {
		fmt.Println("publish:", string(data))
	}
	dc := app.node.DataChannel()
	if dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel not open: %s", dc.ReadyState())
	}
	if dc.BufferedAmount() > 128 && !force {
		return fmt.Errorf("buffer full")
	}
	return dc.Send(data)
}

func (app *Application) Connect(ctx context.Context) error {
	err := app.node.Connect(ctx, app.dest)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) Close() error {
	log.Println("application closed")
	return app.node.Close()
}

func (app *Application) Run() {
	// canvas作成
	canvas := document.Call("createElement", "canvas")
	document.Get("body").Call("appendChild", canvas)

	// Renderer設定
	renderer := THREE.Get("WebGLRenderer").New(map[string]interface{}{
		"canvas":    canvas,
		"antialias": true,
		"alpha":     true,
	})
	app.renderer = renderer
	js.Global().Set("renderer", renderer)
	renderer.Call("setPixelRatio", window.Get("devicePixelRatio"))
	renderer.Call("setSize", window.Get("innerWidth"), window.Get("innerHeight"))

	// シーンとカメラ
	app.scene = THREE.Get("Scene").New()
	app.camera = THREE.Get("Camera").New()
	app.scene.Call("add", app.camera)

	// AR Toolkit初期化
	app.initARContext()

	// レンダリング開始
	app.renderer.Call("setAnimationLoop", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.render()
		return nil
	}))
	// リサイズイベント
	resize := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.onResize()
		return nil
	})
	window.Call("addEventListener", "resize", resize)
	window.Call("addEventListener", "orientationchange", resize)
	app.fire = false
	scope := document.Call("getElementById", "scope")
	document.Get("body").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("click!")
		args[0].Call("preventDefault")
		scope.Get("classList").Call("add", "flash")
		time.AfterFunc(20*time.Millisecond, func() {
			scope.Get("classList").Call("remove", "flash")
		})
		app.fire = true
		return nil
	}))
}

func (app *Application) initARContext() {
	arSource := THREEx.Get("ArToolkitSource").New(map[string]interface{}{
		"sourceType":   "webcam",
		"sourceWidth":  1280,
		"sourceHeight": 720,
	})
	app.arToolkitSrc = arSource
	initCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		/*
			track := arSource.Get("domElement").Call("captureStream").Call("getVideoTracks").Index(0)
			cap := track.Call("getCapabilities")
			console.Call("log", "track:", cap)
		*/
		ctx := THREEx.Get("ArToolkitContext").New(map[string]interface{}{
			"cameraParametersUrl": "camera_para.dat",
			"detectionMode":       "mono",
			"matrixCodeType":      js.Null(),
			"maxDetectionRate":    30,
		})
		app.renderer.Get("domElement").Set("width", arSource.Get("domElement").Get("videoWidth"))
		app.renderer.Get("domElement").Set("height", arSource.Get("domElement").Get("videoHeight"))
		app.arToolkitCtx = ctx
		ctx.Call("init", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// カメラの射影行列更新
			app.camera.Get("projectionMatrix").Call("copy", ctx.Call("getProjectionMatrix"))
			// マーカー作成
			app.createMarkers()
			fmt.Println("AR Initialized")
			//go func() {
			//time.Sleep(1 * time.Second)
			app.onResize()
			//}()
			return nil
		}))
		return nil
	})
	arSource.Call("init", initCallback)
}

func (app *Application) createMarkers() {
	for i, u := range app.patternUrls {
		root := THREE.Get("Group").New()
		app.scene.Call("add", root)

		controls := THREEx.Get("ArMarkerControls").New(
			app.arToolkitCtx,
			root,
			map[string]interface{}{
				"type":             "pattern",
				"patternUrl":       u,
				"changeMatrixMode": "modelViewMatrix",
				"minConfidence":    0.5,
				"smooth":           true,
				"smoothCount":      5,
				"smoothTolerance":  0.005,
				"smoothThreshold":  2,
			},
		)
		controls.Call("addEventListener", "markerFound", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			controls.Set("detected", true)
			return nil
		}))
		controls.Call("addEventListener", "markerLost", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			controls.Set("detected", false)
			return nil
		}))

		// 軸追加
		axes := THREE.Get("AxesHelper").New(0.5)
		root.Call("add", axes)

		app.markers = append(app.markers, root)
		fmt.Printf("marker %d: %s created\n", i, u)
	}
}

func (app *Application) projection(marker js.Value, width, height float64) Marker {
	// marker.matrixWorld から位置を取得
	matrixWorld := marker.Get("matrixWorld")
	pos := js.Global().Get("THREE").Get("Vector3").New()
	pos.Call("setFromMatrixPosition", matrixWorld)

	// camera.project() でNDC座標に変換
	pos.Call("project", app.camera)

	// NDC → スクリーン座標 (Y反転)
	posX := pos.Get("x").Float()
	posY := pos.Get("y").Float()
	screenX := (posX + 1.0) / 2.0 * width
	screenY := (1.0 - posY) / 2.0 * height

	return Marker{
		Point:    Point{X: screenX, Y: screenY},
		Detected: marker.Get("detected").Truthy(),
	}
}

func (app *Application) render() {
	if app.arToolkitSrc.Truthy() && app.arToolkitSrc.Get("ready").Truthy() {
		domElement := app.arToolkitSrc.Get("domElement")
		if domElement.Get("videoWidth").Int() > 0 && domElement.Get("videoHeight").Int() > 0 {
			app.arToolkitCtx.Call("update", domElement)
		}
	}
	w, h := window.Get("innerWidth").Float(), window.Get("innerHeight").Float()
	res := [4]Marker{}
	for i, marker := range app.markers {
		res[i] = app.projection(marker, w, h)
	}
	app.OnUpdate(res)
	app.renderer.Call("render", app.scene, app.camera)
}

func (app *Application) onResize() {
	if !app.arToolkitSrc.Truthy() || !app.arToolkitSrc.Get("ready").Truthy() {
		return
	}
	w, h := window.Get("innerWidth"), window.Get("innerHeight")
	canvas := app.renderer.Get("domElement")
	video := app.arToolkitSrc.Get("domElement")
	video.Get("style").Set("width", w)
	video.Get("style").Set("height", h)
	app.renderer.Call("setSize", w, h)
	app.arToolkitSrc.Call("onResizeElement")
	app.arToolkitSrc.Call("copyElementSizeTo", canvas)

	canvas.Get("style").Set("width", video.Get("style").Get("width"))
	canvas.Get("style").Set("height", video.Get("style").Get("height"))
	canvas.Get("style").Set("margin-left", video.Get("style").Get("margin-left"))
	canvas.Get("style").Set("margin-top", video.Get("style").Get("margin-top"))

	if app.arToolkitCtx.Truthy() {
		app.camera.Get("projectionMatrix").Call("copy", app.arToolkitCtx.Call("getProjectionMatrix"))
	}
}

func main() {
	if GetParam("name") == "" {
		v := window.Call("prompt", "Enter your name")
		if v.Type() == js.TypeString && v.String() != "" {
			SetParam("name", v.String())
		}
	}
	skip := GetParam("skip") != ""
	fmt.Println("wasm instance started: skip =", skip)
	defer fmt.Println("wasm instance ended")
	app := NewApplication()
	defer app.Close()
	cnt := 0
	go func() {
		if !skip {
			connect := false
			for i := 0; i < 3; i++ {
				fmt.Println("connecting:", app.uid)
				err := app.Connect(context.Background())
				if err == nil {
					connect = true
					break
				}
				log.Println(err)
				time.Sleep(5 * time.Second)
			}
			if !connect {
				log.Fatal("failed to connect")
			}
		}
		app.OnUpdate = func(markers [4]Marker) {
			w, h := window.Get("innerWidth").Float(), window.Get("innerHeight").Float()
			points := compensateMarkers(markers)
			x, y := calc(points, w, h)
			println(x, y, "\n",
				points[0].X, points[0].Y, "\n",
				points[1].X, points[1].Y, "\n",
				points[2].X, points[2].Y, "\n",
				points[3].X, points[3].Y,
			)
			if math.IsNaN(x) {
				x = 0.5
			}
			if math.IsNaN(y) {
				y = 0.5
			}
			if cnt%10 == 0 {
				elm := document.Call("getElementById", "message")
				info := fmt.Sprintf("x:%5.2f, y:%5.2f", x, y)
				if elm.Get("innerText").String() != info {
					elm.Set("innerText", info)
				}
			}
			cnt++
			if !skip {
				info := schema.Info{
					ID:   app.uid,
					Name: app.name,
					X:    x,
					Y:    y,
					Fire: app.fire,
				}
				b, _ := json.Marshal(info)
				if err := app.Publish(b, info.Fire); err != nil {
					return
				}
				app.fire = false
			}
		}
		app.Run()
	}()
	select {}
}
