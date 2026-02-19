package ui

import (
	"time"

	"github.com/mokiat/lacking/game"
	"github.com/nobonobo/gun-shooter/schema"
)

type ActiveMember struct {
	Time time.Time
	Info *schema.Info
}

type GlobalState struct {
	Engine      *game.Engine
	ResourceSet *game.ResourceSet
	Actives     map[string]ActiveMember
}
