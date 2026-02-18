package widget

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/std"
)

var Loading = co.Define[*loadingComponent]()

type loadingComponent struct {
	co.BaseComponent

	elapsedTime   time.Duration
	loadingLabels [][]rune

	font     *ui.Font
	fontSize float32

	maxLabelSize sprec.Vec2
}

func (c *loadingComponent) OnCreate() {
	c.elapsedTime = 0
	c.loadingLabels = [][]rune{
		[]rune("Loading"),
		[]rune("Loading."),
		[]rune("Loading.."),
		[]rune("Loading..."),
	}

	c.font = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
	c.fontSize = 48.0

	lastLabel := c.loadingLabels[len(c.loadingLabels)-1]
	c.maxLabelSize = sprec.Vec2{
		X: c.font.LineWidth(lastLabel, c.fontSize),
		Y: c.font.LineHeight(c.fontSize),
	}
}

func (c *loadingComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithData(std.ElementData{
			Essence:   c,
			IdealSize: opt.V(ui.NewSize(int(c.maxLabelSize.X), int(c.maxLabelSize.Y))),
		})
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithChildren(c.Properties().Children())
	})
}

func (c *loadingComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	c.elapsedTime += canvas.ElapsedTime()

	tickEvery := 500 * time.Millisecond
	tickIndex := int(c.elapsedTime / tickEvery)
	text := c.loadingLabels[tickIndex%len(c.loadingLabels)]

	drawBounds := canvas.DrawBounds(element, false)

	canvas.Push()
	canvas.Translate(drawBounds.Position)
	canvas.Translate(sprec.Vec2{
		X: (drawBounds.Size.X - c.maxLabelSize.X) / 2,
		Y: (drawBounds.Size.Y - c.maxLabelSize.Y) / 2,
	})
	canvas.FillTextLine(text, sprec.ZeroVec2(), ui.Typography{
		Font:  c.font,
		Size:  c.fontSize,
		Color: ui.White(),
	})
	canvas.Pop()

	element.Invalidate() // force redraw
}
