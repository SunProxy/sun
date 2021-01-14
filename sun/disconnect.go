package sun

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type PlanetDisconnect struct {
	Message string
}

func (p *PlanetDisconnect) ID() uint32 {
	return IDPlanetDisconnect
}

func (p *PlanetDisconnect) Marshal(w *protocol.Writer) {
	w.String(&p.Message)
}

func (p *PlanetDisconnect) Unmarshal(r *protocol.Reader) {
	r.String(&p.Message)
}

