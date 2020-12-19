package sun

import "github.com/sandertv/gophertunnel/minecraft/protocol"

/*
Text is sent by the server to send a message to all the connected players on the proxy.
*/
type Text struct {
	Message string
}

func (pk Text) ID() uint32 {
	return IDSunText
}

func (pk Text) Marshal(w *protocol.Writer) {
	w.String(&pk.Message)
}

func (pk Text) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Message)
}
