package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"syscall/js"

	"github.com/google/uuid"
	"github.com/nobonobo/rtcconnect/node"
	"github.com/pion/webrtc/v4"
)

type Info struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Fire bool    `json:"fire"`
}

var (
	instance *node.Node
	stop     = func() {}
	mu       sync.RWMutex
	informs  = map[string]*Info{}
)

func UUID() string {
	uid, _ := uuid.NewV6()
	return uid.String()
}

func Setup(id string) {
	stop()
	log.Println("Setup:", id)
	instance = node.New(id)
}

func ID() string {
	return instance.ID()
}

func CLose() {
	mu.Lock()
	defer mu.Unlock()
	stop()
	instance.Close()
	instance = nil
	stop = func() {}
}

func Listen(hostID string) {
	stop()
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	go func() {
		informs = map[string]*Info{}
		instance = node.New(hostID)
		log.Println("Listening on:", hostID)
		instance.OnConnected = func(n *node.Node) {
			id := n.ID()
			log.Println("Connected:", id)
			n.PeerConnection().OnDataChannel(func(dc *webrtc.DataChannel) {
				log.Println("DataChannel:", id)
				dc.OnMessage(func(msg webrtc.DataChannelMessage) {
					var info *Info
					if err := json.Unmarshal(msg.Data, &info); err != nil {
						log.Println(err)
						return
					}
					mu.Lock()
					informs[info.ID] = info
					mu.Unlock()
				})
				dc.OnClose(func() {
					mu.Lock()
					delete(informs, id)
					mu.Unlock()
				})
			})
		}
		if err := instance.Listen(ctx); err != nil {
			log.Println(err)
		}
	}()
}

func Inform() string {
	mu.RLock()
	defer mu.RUnlock()
	b, _ := json.Marshal(informs)
	return string(b)
}

func Stop() {
	stop()
}

func Connect(name, dst string) {
	stop()
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	go func() {
		if err := instance.Connect(ctx, dst); err != nil {
			log.Println(err)
		}
		stop = func() {
			cancel()
		}
		instance.DataChannel().OnOpen(func() {
			log.Println("DataChannel opened:", name)
			Send(name, 0, 0, false)
		})
	}()
}

func Send(name string, x, y float64, fire bool) {
	if instance != nil {
		msg := Info{
			ID:   instance.ID(),
			Name: name,
			X:    x,
			Y:    y,
			Fire: fire,
		}
		b, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			return
		}
		if err := instance.DataChannel().Send(b); err != nil {
			log.Println(err)
		}
	}
}

func main() {
	js.Global().Set("Go", js.ValueOf(map[string]interface{}{
		"UUID": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return UUID()
		}),
		"Setup": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			Setup(args[0].String())
			return nil
		}),
		"ID": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return ID()
		}),
		"Listen": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			Listen(args[0].String())
			return nil
		}),
		"Stop": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			Stop()
			return nil
		}),
		"Connect": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			Connect(args[0].String(), args[1].String())
			return nil
		}),
		"Inform": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return Inform()
		}),
		"Send": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			Send(args[0].String(), args[1].Float(), args[2].Float(), args[3].Bool())
			return nil
		}),
	}))
	log.Println("WASM loaded")
	select {}
}
