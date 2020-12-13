package sun

import "github.com/sandertv/gophertunnel/minecraft"

type Sun struct {
	listener *minecraft.Listener
	players []*Player
}