package widget

import (
	"bytes"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	std "github.com/mokiat/lacking/ui/std"
	"github.com/skip2/go-qrcode"
)

var QRCode = co.Define[*qrCodeComponent]()

type QRCodeData struct {
	Text string
	Size float32
}

var defaultQRCodeData = QRCodeData{
	Text: "",
	Size: 128,
}

type QRCodeCallbackData struct {
	OnClick std.OnActionFunc
}

var defaultQRCodeCallbackData = QRCodeCallbackData{
	OnClick: func() {},
}

type qrCodeComponent struct {
	co.BaseComponent

	data    QRCodeData
	qrImage *ui.Image
	text    string
}

func (c *qrCodeComponent) OnUpsert() {
	data := co.GetOptionalData(c.Properties(), defaultQRCodeData)

	c.data = data
	c.text = data.Text
	c.updateQRImage()
}

func (c *qrCodeComponent) updateQRImage() {
	if c.data.Text == "" {
		c.qrImage = nil
		return
	}
	qr, err := qrcode.New(c.data.Text, qrcode.Medium)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	if err := qr.Write(int(c.data.Size), &buf); err != nil {
		return
	}
	ctx := c.Scope().Context()
	img, err := ctx.CreateImage(qr.Image(int(c.data.Size)))
	if err != nil {
		return
	}
	c.qrImage = img
}

func (c *qrCodeComponent) Render() co.Instance {
	// === 方法3: Render内でElement経由 ===
	padding := ui.Spacing{Left: 5, Right: 5, Top: 5, Bottom: 5}

	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Essence:   c,
			Padding:   padding,
			IdealSize: opt.V(ui.NewSize(int(c.data.Size), int(c.data.Size))),
		})
		co.WithChildren(c.Properties().Children())
	})
}

func (c *qrCodeComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	drawBounds := canvas.DrawBounds(element, false)
	canvas.Reset()
	canvas.Rectangle(
		drawBounds.Position,
		drawBounds.Size,
	)
	canvas.Fill(ui.Fill{
		Rule:        ui.FillRuleSimple,
		Color:       ui.White(),
		Image:       c.qrImage,
		ImageOffset: drawBounds.Position,
		ImageSize:   drawBounds.Size,
	})
	element.Invalidate() // force redraw
}

// 使用例
// widget.NewQRCode().
//     WithData(widget.QRCodeData{Text: "https://example.com", Size: 200}).
//     WithCallbackData(widget.QRCodeCallbackData{
//         OnClick: func() { fmt.Println("QR clicked!") },
//     })
