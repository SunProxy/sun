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

package planet

import (
	"bytes"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"log"
	"net"
)

type Planet struct {
	buf  bytes.Buffer
	pool packet.Pool
	conn net.Conn
	Id   uuid.UUID
}

//Used to create a NewPlanet
//noinspection GoUnusedFunction
func NewPlanet(conn net.Conn) *Planet {
	return &Planet{conn: conn, pool: newTcpPool()}
}

func newTcpPool() packet.Pool {
	pool := make(packet.Pool)
	pool[sunpacket.IDPlanetAuth] = &sunpacket.PlanetAuth{}
	pool[sunpacket.IDPlanetDisconnect] = &sunpacket.PlanetDisconnect{}
	pool[sunpacket.IDPlanetTransfer] = &sunpacket.PlanetTransfer{}
	pool[sunpacket.IDPlanetTransferResponse] = &sunpacket.TransferResponse{}
	pool[sunpacket.IDPlanetText] = &sunpacket.Text{}
	pool[sunpacket.IDPlanetTextResponse] = &sunpacket.TextResponse{}
	pool[sunpacket.IDPlanetUserInfo] = &sunpacket.UserInfo{}
	return pool
}

func (p *Planet) ReadPacket() (packet.Packet, error) {
	var length uint32
	err := binary.Read(p.conn, binary.LittleEndian, &length)
	if err != nil {
		log.Printf("Invalid packet recieved length not sent from planet: %s!\n", p.conn.RemoteAddr())
		return nil, err
	}
	var id uint32
	err = binary.Read(p.conn, binary.LittleEndian, &id)
	if err != nil {
		log.Printf("Invalid packet recieved didn't send correct packet formated from planet: %s!\n", p.conn.RemoteAddr())
		return nil, err
	}
	return p.pool[id], err
}

func (p *Planet) WritePacket(pk packet.Packet) error {
	pk.Marshal(protocol.NewWriter(&p.buf, 0))
	buf := bytes.NewBuffer(make([]byte, 0, 4+len(p.buf.Bytes())))
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(p.buf.Bytes()))); err != nil {
		return err
	}
	if _, err := buf.Write(p.buf.Bytes()); err != nil {
		return err
	}

	if _, err := p.conn.Write(buf.Bytes()); err != nil {
		return err
	}
	p.buf.Reset()
	return nil
}

func (p *Planet) Conn() net.Conn {
	return p.conn
}
