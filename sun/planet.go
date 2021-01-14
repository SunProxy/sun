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
	"bytes"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"log"
	"net"
)

type Planet struct {
	buf bytes.Buffer
	pool packet.Pool
	conn net.Conn
	id uuid.UUID
}

func NewPlanet(ip IpAddr) (*Planet, error) {
	conn, err := net.Dial("tcp", ip.ToString())
	if err != nil {
		return &Planet{}, err
	}
	return &Planet{conn: conn}, nil
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

func (s *Sun) handlePlanet(planet *Planet) {
	go func() {
		for {
			pk, err := planet.ReadPacket()
			if err != nil {
				log.Println(err)
				return
			}
			if pk, ok := pk.(*PlanetTransfer); ok {
				if ray, ok := s.Rays[pk.User]; ok {
					s.TransferRay(ray, IpAddr{Address: pk.Address, Port: pk.Port})
				} else {
					log.Printf("Received bad request from planet: %s, the player by uuid %s was not found!\n", planet.conn.RemoteAddr(), pk.User)
				}
				continue
			}
			if pk, ok := pk.(*Text); ok {
				//Only iterate if we have to.
				if len(pk.Servers) > 0 {
					//in a new routine because of the iteration
					go s.SendMessageToServers(pk.Message, pk.Servers)
					continue
				}
				//if len(pk.Servers) == 0
				s.SendMessage(pk.Message)
				continue
			}
		}
	}()
}

