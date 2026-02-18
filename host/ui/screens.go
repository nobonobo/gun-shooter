package ui

import (
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/std"
	"github.com/mokiat/lacking/util/async"

	"github.com/nobonobo/gun-shooter/host/resources"
	"github.com/nobonobo/gun-shooter/host/ui/widget"
)

// Global state variables for UI
var (
	loadingState LoadingState
	loadingError error
)

// --- Intro Screen ---

var IntroScreen = co.Define[*introScreenComponent]()

type IntroScreenData struct {
	App *applicationComponent
}

type introScreenComponent struct {
	co.BaseComponent
}

func (c *introScreenComponent) OnCreate() {
	co.Window(c.Scope()).SetCursorVisible(false)

	globalState := co.TypedValue[GlobalState](c.Scope())
	engine := globalState.Engine
	resourceSet := globalState.ResourceSet

	componentData := co.GetData[IntroScreenData](c.Properties())
	app := componentData.App

	promise := NewLoadingPromise(
		co.Window(c.Scope()),
		LoadHomeData(engine, resourceSet),
		func(d *HomeData) {
			homeSceneData = d
		},
		func(err error) {
			loadingError = err
		},
	)
	loadingState = LoadingState{
		Promise:         promise,
		SuccessViewName: ViewNameHome,
		ErrorViewName:   ViewNameError,
	}

	co.After(c.Scope(), time.Second, func() {
		app.SetActiveView(ViewNameLoading)
	})
}

func (c *introScreenComponent) OnDelete() {
	co.Window(c.Scope()).SetCursorVisible(true)
}

func (c *introScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("logo-picture", co.New(std.Picture, func() {
			co.WithLayoutData(layout.Data{
				Width:            opt.V(512),
				Height:           opt.V(128),
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(std.PictureData{
				BackgroundColor: opt.V(ui.Transparent()),
				Image:           co.OpenImage(c.Scope(), "ui/images/logo.png"),
				Mode:            std.ImageModeFit,
			})
		}))
	})
}

// --- Loading Screen ---

var LoadingScreen = co.Define[*loadingScreenComponent]()

type LoadingScreenData struct {
	App *applicationComponent
}

type loadingScreenComponent struct {
	co.BaseComponent
}

func (c *loadingScreenComponent) OnCreate() {
	componentData := co.GetData[LoadingScreenData](c.Properties())
	app := componentData.App

	state := loadingState
	state.Promise.OnSuccess(func() {
		app.SetActiveView(state.SuccessViewName)
	})
	state.Promise.OnError(func() {
		app.SetActiveView(state.ErrorViewName)
	})
}

func (c *loadingScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("loading", co.New(widget.Loading, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
		}))
	})
}

type LoadingState struct {
	Promise         LoadingPromise
	SuccessViewName ViewName
	ErrorViewName   ViewName
}

type LoadingPromise interface {
	OnSuccess(func())
	OnError(func())
}

func NewLoadingPromise[T any](worker game.Worker, promise async.Promise[T], onSuccess func(T), onError func(error)) LoadingPromise {
	return &loadingPromise[T]{
		worker:    worker,
		promise:   promise,
		onSuccess: onSuccess,
		onError:   onError,
	}
}

type loadingPromise[T any] struct {
	worker    game.Worker
	promise   async.Promise[T]
	onSuccess func(T)
	onError   func(error)
}

func (p *loadingPromise[T]) OnSuccess(cb func()) {
	p.promise.OnSuccess(func(value T) {
		p.worker.Schedule(func() {
			p.onSuccess(value)
			cb()
		})
	})
}

func (p *loadingPromise[T]) OnError(cb func()) {
	p.promise.OnError(func(err error) {
		p.worker.Schedule(func() {
			p.onError(err)
			cb()
		})
	})
}

// --- Error Screen ---

var ErrorScreen = co.Define[*errorScreenComponent]()

type ErrorScreenData struct {
	App *applicationComponent
}

var _ ui.ElementKeyboardHandler = (*errorScreenComponent)(nil)

type errorScreenComponent struct {
	co.BaseComponent

	titleFont     *ui.Font
	titleFontSize float32

	messageFont     *ui.Font
	messageFontSize float32

	message string
}

func (c *errorScreenComponent) OnCreate() {
	c.message = c.formatError(loadingError)

	c.titleFont = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
	c.titleFontSize = float32(48.0)

	c.messageFont = co.OpenFont(c.Scope(), "ui:///roboto-regular.ttf")
	c.messageFontSize = float32(24.0)
}

func (c *errorScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("handler", co.New(std.Element, func() {
			co.WithLayoutData(layout.Data{
				Left:   opt.V(0),
				Right:  opt.V(0),
				Top:    opt.V(0),
				Bottom: opt.V(0),
			})
			co.WithData(std.ElementData{
				Essence:       c,
				Enabled:       opt.V(true),
				CanAutoFocus:  opt.V(true),
				CreateFocused: true,
			})
		}))

		co.WithChild("title", co.New(std.Label, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(-150),
			})
			co.WithData(std.LabelData{
				Text:      "ERROR",
				Font:      c.titleFont,
				FontSize:  opt.V(c.titleFontSize),
				FontColor: opt.V(ui.White()),
			})
		}))

		co.WithChild("info", co.New(std.Label, func() {
			co.WithLayoutData(layout.Data{
				HorizontalCenter: opt.V(0),
				VerticalCenter:   opt.V(0),
			})
			co.WithData(std.LabelData{
				Text:      c.message,
				Font:      c.messageFont,
				FontSize:  opt.V(c.messageFontSize),
				FontColor: opt.V(ui.White()),
			})
		}))
	})
}

func (c *errorScreenComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	if event.Action == ui.KeyboardActionUp && event.Code == ui.KeyCodeEscape {
		co.Window(c.Scope()).Close()
	}
	return true
}

func (c *errorScreenComponent) formatError(err error) string {
	wordWrap := func(text string, maxLineLength int) iter.Seq[string] {
		return func(yield func(string) bool) {
			runes := []rune(text)
			for len(runes) > maxLineLength {
				if !yield(string(runes[:maxLineLength])) {
					return
				}
				runes = runes[maxLineLength:]
			}
			if !yield(string(runes)) {
				return
			}
		}
	}

	var builder strings.Builder
	fmt.Fprintln(&builder, "The game has encountered an error. Press ESCAPE to exit.")
	fmt.Fprintln(&builder)
	fmt.Fprint(&builder, "Error: ")
	for line := range wordWrap(err.Error(), 80) {
		fmt.Fprintln(&builder, line)
	}
	return builder.String()
}

// --- Licenses Screen ---

var LicensesScreen = co.Define[*licensesScreenComponent]()

type LicensesScreenData struct {
	App *applicationComponent
}

type licensesScreenComponent struct {
	co.BaseComponent

	app *applicationComponent
}

func (c *licensesScreenComponent) OnCreate() {
	componentData := co.GetData[LicensesScreenData](c.Properties())
	c.app = componentData.App
}

func (c *licensesScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		co.WithChild("menu-pane", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Bottom: opt.V(0),
				Left:   opt.V(0),
				Width:  opt.V(200),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.Black()),
				Layout:          layout.Anchor(),
			})

			co.WithChild("button", co.New(widget.Button, func() {
				co.WithLayoutData(layout.Data{
					HorizontalCenter: opt.V(0),
					Bottom:           opt.V(100),
				})
				co.WithData(widget.ButtonData{
					Text: "Back",
				})
				co.WithCallbackData(widget.ButtonCallbackData{
					OnClick: c.onBackClicked,
				})
			}))
		}))

		co.WithChild("content-pane", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Bottom: opt.V(0),
				Left:   opt.V(200),
				Right:  opt.V(0),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.RGB(0x11, 0x11, 0x11)),
				Layout:          layout.Anchor(),
			})

			co.WithChild("title", co.New(std.Label, func() {
				co.WithLayoutData(layout.Data{
					Top:              opt.V(15),
					Height:           opt.V(32),
					HorizontalCenter: opt.V(0),
				})
				co.WithData(std.LabelData{
					Font:      co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf"),
					FontSize:  opt.V(float32(32)),
					FontColor: opt.V(ui.White()),
					Text:      "Open-Source Licenses",
				})
			}))

			co.WithChild("sub-title", co.New(std.Label, func() {
				co.WithLayoutData(layout.Data{
					Top:              opt.V(50),
					Height:           opt.V(20),
					HorizontalCenter: opt.V(0),
				})
				co.WithData(std.LabelData{
					Font:      co.OpenFont(c.Scope(), "ui:///roboto-italic.ttf"),
					FontSize:  opt.V(float32(20)),
					FontColor: opt.V(ui.White()),
					Text:      "- scroll to view all -",
				})
			}))

			co.WithChild("license-scroll-pane", co.New(std.ScrollPane, func() {
				co.WithLayoutData(layout.Data{
					Top:    opt.V(80),
					Bottom: opt.V(0),
					Left:   opt.V(0),
					Right:  opt.V(0),
				})
				co.WithData(std.ScrollPaneData{
					DisableHorizontal: true,
					DisableVertical:   false,
					CreateFocused:     true,
				})

				co.WithChild("license-holder", co.New(std.Element, func() {
					co.WithLayoutData(layout.Data{
						GrowHorizontally: true,
					})
					co.WithData(std.ElementData{
						Padding: ui.Spacing{
							Top:    100,
							Bottom: 100,
						},
						Layout: layout.Anchor(),
					})

					co.WithChild("license-text", co.New(std.Label, func() {
						co.WithLayoutData(layout.Data{
							HorizontalCenter: opt.V(0),
							VerticalCenter:   opt.V(0),
						})
						co.WithData(std.LabelData{
							Font:      co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf"),
							FontSize:  opt.V(float32(16)),
							FontColor: opt.V(ui.White()),
							Text:      resources.Licenses,
						})
					}))
				}))
			}))
		}))
	})
}

func (c *licensesScreenComponent) onBackClicked() {
	c.app.SetActiveView(ViewNameHome)
}
