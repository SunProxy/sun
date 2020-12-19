package sun

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

//Transfer is sent by the server to change a Players remote connection otherwise known as the fast transfer packet
type Transfer struct {
	// Address is the address of the new server, which might be either a hostname or an actual IP address.
	Address string
	// Port is the UDP port of the new server.
	Port uint16
}

func (pk Transfer) ID() uint32 {
	return IDSunTransfer
}

func (pk Transfer) Marshal(w *protocol.Writer) {
	w.String(&pk.Address)
	w.Uint16(&pk.Port)
}

func (pk Transfer) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Address)
	r.Uint16(&pk.Port)
}
