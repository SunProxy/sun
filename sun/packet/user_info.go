package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type UserInfo struct {
	Uuid   string
	Xuid   string
	IpAddr string
}

func (u UserInfo) ID() uint32 {
	return IDPlanetUserInfo
}

func (u UserInfo) Marshal(w *protocol.Writer) {
	w.String(&u.Uuid)
	w.String(&u.Xuid)
	w.String(&u.IpAddr)
}

func (u UserInfo) Unmarshal(r *protocol.Reader) {
	r.String(&u.Uuid)
	r.String(&u.Xuid)
	r.String(&u.IpAddr)
}
