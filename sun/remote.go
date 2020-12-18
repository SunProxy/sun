package sun

import (
	"github.com/sandertv/gophertunnel/minecraft"
)

type Remote struct {
	conn *minecraft.Conn
	addr IpAddr
}
