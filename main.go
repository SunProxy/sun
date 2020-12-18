package main

import (
	"github.com/jviguy/sun/sun"
	"github.com/sandertv/gophertunnel/minecraft"
)

func main() {
	listener, _ := minecraft.Listen("raknet", ":19132")
	s := sun.Sun{Hub: sun.IpAddr{Ip: "0.0.0.0", Port: 19133}, Listener: listener, Players: make(map[string]*sun.Player)}
	s.Start()
}
