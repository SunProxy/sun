/**
      ___           ___           ___
     /  /\         /__/\         /__/\
    /  /:/_        \  \:\        \  \:\
   /  /:/ /\        \  \:\        \  \:\
  /  /:/ /::\   ___  \  \:\   _____\__\:\
 /__/:/ /:/\:\ /__/\  \__\:\ /__/::::::::\
 \  \:\/:/~/:/ \  \:\ /  /:/ \  \:\~~\~~\/
  \  \::/ /:/   \  \:\  /:/   \  \:\  ~~~
   \__\/ /:/     \  \:\/:/     \  \:\
     /__/:/       \  \::/       \  \:\
     \__\/         \__\/         \__\/

MIT License

Copyright (c) 2020 Jviguy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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

func (pk *Transfer) ID() uint32 {
	return IDRayTransfer
}

func (pk *Transfer) Marshal(w *protocol.Writer) {
	w.String(&pk.Address)
	w.Uint16(&pk.Port)
}

func (pk *Transfer) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Address)
	r.Uint16(&pk.Port)
}

type PlanetTransfer struct {
	// Address is the address of the new server, which might be either a hostname or an actual IP address.
	Address string
	// Port is the UDP port of the new server.
	Port uint16
	//User is the uuid of the given player to transfer
	User string
}

func (pk *PlanetTransfer) ID() uint32 {
	return IDPlanetTransfer
}

func (pk *PlanetTransfer) Marshal(w *protocol.Writer) {
	w.String(&pk.Address)
	w.Uint16(&pk.Port)
	w.String(&pk.User)
}

func (pk *PlanetTransfer) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Address)
	r.Uint16(&pk.Port)
	r.String(&pk.User)
}