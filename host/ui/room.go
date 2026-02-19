package ui

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/layout"
	"github.com/mokiat/lacking/ui/mvc"
	"github.com/mokiat/lacking/ui/std"
	"github.com/pion/webrtc/v4"

	"github.com/nobonobo/gun-shooter/host/ui/widget"
	"github.com/nobonobo/gun-shooter/schema"
	"github.com/nobonobo/rtcconnect/node"
)

var RoomScreen = mvc.EventListener(co.Define[*roomScreenComponent]())

type RoomMembersUpdatedEvent struct {
	Members []string
}

type RoomScreenData struct {
	App *applicationComponent
}

type roomScreenComponent struct {
	co.BaseComponent

	app *applicationComponent

	engine      *game.Engine
	resourceSet *game.ResourceSet

	titleFont *ui.Font
	textFont  *ui.Font
	actives   map[string]struct {
		time time.Time
		name string
	}
	members []string
	host    *node.Node
	ctx     context.Context
	cancel  func()
}

func (c *roomScreenComponent) OnCreate() {
	globalState := co.TypedValue[GlobalState](c.Scope())
	c.engine = globalState.Engine
	c.resourceSet = globalState.ResourceSet

	componentData := co.GetData[RoomScreenData](c.Properties())
	c.app = componentData.App

	c.titleFont = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
	c.textFont = co.OpenFont(c.Scope(), "ui:///roboto-regular.ttf")
	c.host = node.NewHost(GetParam("id"))
	c.actives = map[string]struct {
		time time.Time
		name string
	}{}
	eventBus := co.TypedValue[*mvc.EventBus](c.Scope())
	c.host.OnConnected = func(peer *node.Node) {
		id := peer.ID()
		peer.PeerConnection().OnDataChannel(func(dc *webrtc.DataChannel) {
			log.Println("data channel opened:", id)
			dc.OnClose(func() {
				log.Println("data channel closed:", id)
			})
			cnt := 0
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				var info *schema.Info
				if err := json.Unmarshal(msg.Data, &info); err != nil {
					log.Println("failed to unmarshal info:", err)
					return
				}
				cnt++
				if cnt%1000 == 0 {
					log.Println("data channel message:", id, info)
				}
				c.actives[id] = struct {
					time time.Time
					name string
				}{
					time: time.Now(),
					name: info.Name,
				}
				members := []string{}
				for _, active := range c.actives {
					if time.Since(active.time) > 5*time.Second {
						continue
					}
					members = append(members, active.name)
				}
				eventBus.Notify(RoomMembersUpdatedEvent{Members: members})
			})
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	go func() {
		log.Println("listen start:", c.host.ID())
		defer log.Println("listen stop:", c.host.ID())
		if err := c.host.Listen(ctx); err != nil {
			log.Println("failed to listen", err)
		}
	}()
}

func (c *roomScreenComponent) OnDelete() {
	c.cancel()
}

func (c *roomScreenComponent) Render() co.Instance {
	return co.New(std.Container, func() {
		co.WithData(std.ContainerData{
			BackgroundColor: opt.V(ui.Black()),
			Layout:          layout.Anchor(),
		})

		// Left Pane: Info and Back button
		co.WithChild("menu-pane", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Bottom: opt.V(0),
				Left:   opt.V(0),
				Width:  opt.V(320),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.Black()),
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

				co.WithChild("back-button", co.New(widget.Button, func() {
					co.WithData(widget.ButtonData{
						Text: "Back",
					})
					co.WithCallbackData(widget.ButtonCallbackData{
						OnClick: c.onBackClicked,
					})
				}))
			}))
		}))

		// Content Pane
		co.WithChild("content-pane", co.New(std.Container, func() {
			co.WithLayoutData(layout.Data{
				Top:    opt.V(0),
				Bottom: opt.V(0),
				Left:   opt.V(320),
				Right:  opt.V(0),
			})
			co.WithData(std.ContainerData{
				BackgroundColor: opt.V(ui.RGB(0x11, 0x11, 0x11)),
				Layout:          layout.Anchor(),
			})

			// QR Code Section
			co.WithChild("qr-section", co.New(std.Element, func() {
				co.WithLayoutData(layout.Data{
					Top:              opt.V(20),
					HorizontalCenter: opt.V(-170),
					Width:            opt.V(320),
					Height:           opt.V(320),
				})
				co.WithData(std.ElementData{
					Layout: layout.Anchor(),
				})
				link := BaseURL() + "scope/?dest=" + GetParam("id")
				co.WithChild("qr-code", co.New(widget.QRCode, func() {
					co.WithLayoutData(layout.Data{
						HorizontalCenter: opt.V(0),
						VerticalCenter:   opt.V(0),
					})
					co.WithData(widget.QRCodeData{
						Text: link,
						Size: 320,
					})
					co.WithCallbackData(widget.QRCodeCallbackData{
						OnClick: func() {
							log.Println("QR Code clicked:", link)
							URLOpen(link)
						},
					})
				}))
			}))

			// Members List Section
			co.WithChild("members-section", co.New(std.Container, func() {
				co.WithLayoutData(layout.Data{
					Top:              opt.V(20),
					HorizontalCenter: opt.V(170),
					Width:            opt.V(320),
					Height:           opt.V(320),
				})
				co.WithData(std.ContainerData{
					Layout: layout.Vertical(layout.VerticalSettings{
						ContentAlignment: layout.HorizontalAlignmentCenter,
						ContentSpacing:   10,
					}),
				})

				co.WithChild("members-title", co.New(std.Label, func() {
					co.WithData(std.LabelData{
						Font:      c.titleFont,
						FontSize:  opt.V(float32(24)),
						FontColor: opt.V(ui.Black()),
						Text:      "Members:",
					})
				}))

				// Dynamic Members
				for _, name := range c.members {
					co.WithChild("member-"+name, co.New(std.Label, func() {
						co.WithData(std.LabelData{
							Font:      c.textFont,
							FontSize:  opt.V(float32(20)),
							FontColor: opt.V(ui.RGB(0xAA, 0xAA, 0xAA)),
							Text:      name,
						})
					}))
				}
			}))
		}))
	})
}

func (c *roomScreenComponent) OnEvent(event mvc.Event) {
	switch e := event.(type) {
	case RoomMembersUpdatedEvent:
		c.members = e.Members
		c.Invalidate() // 再描画を要求
	}
}

func (c *roomScreenComponent) onPlayClicked() {
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
	log.Println("play clicked from room")
}

func (c *roomScreenComponent) onBackClicked() {
	c.app.SetActiveView(ViewNameHome)
}
