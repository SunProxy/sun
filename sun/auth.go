package sun

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type PlanetAuth struct {
	Key string
}

func (p *PlanetAuth) ID() uint32 {
	return IDPlanetAuth
}

func (p *PlanetAuth) Marshal(w *protocol.Writer) {
	w.String(&p.Key)
}

func (p *PlanetAuth) Unmarshal(r *protocol.Reader) {
	r.String(&p.Key)
}

