package ui

import (
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/ui/std"
)

func BootstrapApplication(window *ui.Window, gameController *game.Controller) {
	engine := gameController.Engine()
	eventBus := mvc.NewEventBus()

	scope := co.RootScope(window)
	scope = co.TypedValueScope(scope, eventBus)
	scope = co.TypedValueScope(scope, GlobalState{
		Engine:      engine,
		ResourceSet: engine.CreateResourceSet(),
	})
	co.Initialize(scope, co.New(Application, nil))
}

var Application = mvc.EventListener(co.Define[*applicationComponent]())

type applicationComponent struct {
	co.BaseComponent

	eventBus   *mvc.EventBus
	activeView ViewName
}

func (c *applicationComponent) OnCreate() {
	c.eventBus = co.TypedValue[*mvc.EventBus](c.Scope())
	c.activeView = ViewNameIntro
}

func (c *applicationComponent) Render() co.Instance {
	return co.New(std.Switch, func() {
		co.WithData(std.SwitchData{
			ChildKey: c.activeView,
		})

		co.WithChild(ViewNameIntro, co.New(IntroScreen, func() {
			co.WithData(IntroScreenData{
				App: c,
			})
		}))
		co.WithChild(ViewNameError, co.New(ErrorScreen, func() {
			co.WithData(ErrorScreenData{
				App: c,
			})
		}))
		co.WithChild(ViewNameLoading, co.New(LoadingScreen, func() {
			co.WithData(LoadingScreenData{
				App: c,
			})
		}))
		co.WithChild(ViewNameLicenses, co.New(LicensesScreen, func() {
			co.WithData(LicensesScreenData{
				App: c,
			})
		}))
		co.WithChild(ViewNameHome, co.New(HomeScreen, func() {
			co.WithData(HomeScreenData{
				App: c,
			})
		}))
		co.WithChild(ViewNamePlay, co.New(PlayScreen, func() {
			co.WithData(PlayScreenData{
				App: c,
			})
		}))
	})
}

func (c *applicationComponent) OnEvent(event mvc.Event) {
	switch event.(type) {
	case ApplicationActiveViewChangedEvent:
		c.Invalidate()
	}
}

func (c *applicationComponent) ActiveView() ViewName {
	return c.activeView
}

func (c *applicationComponent) SetActiveView(view ViewName) {
	c.activeView = view
	c.eventBus.Notify(ApplicationActiveViewChangedEvent{
		ActiveView: view,
	})
}

const (
	ViewNameIntro    ViewName = "intro"
	ViewNameError    ViewName = "error"
	ViewNameLoading  ViewName = "loading"
	ViewNameLicenses ViewName = "licenses"
	ViewNameHome     ViewName = "home"
	ViewNamePlay     ViewName = "play"
)

type ViewName = string

type ApplicationActiveViewChangedEvent struct {
	ActiveView ViewName
}
