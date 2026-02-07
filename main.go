package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/nobonobo/rtcconnect/node"
	"github.com/pion/webrtc/v4"
)

var (
	instance  *node.Node
	selfID, _ = uuid.NewV6()
)

//go:wasmexport Listen
func Listen() {
	if instance == nil {
		instance = node.New(selfID.String())
	}
	instance.OnConnected = func(n *node.Node) {
		log.Println("Connected:", n.ID())
		n.PeerConnection().OnDataChannel(func(dc *webrtc.DataChannel) {
			log.Println("DataChannel:", n.ID())
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				log.Println("Message:", n.ID(), string(msg.Data))
			})
		})
	}
	if err := instance.Listen(context.Background()); err != nil {
		log.Println(err)
	}
}

//go:wasmexport Connect
func Connect(dst string) {
	if instance == nil {
		instance = node.New(selfID.String())
	}
	if err := instance.Connect(context.Background(), dst); err != nil {
		log.Println(err)
	}
}

func main() {
	log.Println("WASM started")
	select {}
}
