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
	"github.com/sunproxy/sun/sun/ip_addr"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"github.com/sunproxy/sun/sun/planet"
	"log"
	"strings"
)

func (s *Sun) handlePlanet(planet *planet.Planet) {
	go func() {
		for {
			pk, err := planet.ReadPacket()
			if err != nil {
				log.Println(err)
				return
			}
			if pk, ok := pk.(*sunpacket.PlanetTransfer); ok {
				if ray, ok := s.Rays[pk.User]; ok {
					err := s.TransferRay(ray, ip_addr.IpAddr{Address: pk.Address, Port: pk.Port})
					if err != nil {
						if strings.Contains(err.Error(), "no such host") {
							_ = planet.WritePacket(&sunpacket.TransferResponse{Type: sunpacket.TransferResponseRemoteNotFound, Message: err.Error()})
						} else {
							_ = planet.WritePacket(&sunpacket.TransferResponse{Type: sunpacket.TransferResponseRemoteRejection, Message: err.Error()})
						}
						continue
					}
					_ = planet.WritePacket(&sunpacket.TransferResponse{Type: sunpacket.TransferResponseSuccess, Message: "Transfer successful!"})
				} else {
					_ = planet.WritePacket(&sunpacket.TransferResponse{Type: sunpacket.TransferResponseBadRequest, Message: "Player not found in this proxy!"})
					log.Printf("Received bad request from planet: %s, the player by uuid %s was not found!\n", planet.Conn().RemoteAddr(), pk.User)
				}
				continue
			}
			if pk, ok := pk.(*sunpacket.Text); ok {
				if pk.Message == "" {
					_ = planet.WritePacket(&sunpacket.TextResponse{Type: sunpacket.TextResponseBadRequest, Message: "A Message mustn't be empty."})
					continue
				}
				//Only iterate if we have to.
				if len(pk.Servers) > 0 {
					//in a new routine because of the iteration
					go s.SendMessageToServers(pk.Message, pk.Servers)
					_ = planet.WritePacket(&sunpacket.TextResponse{Type: sunpacket.TextResponseSuccess, Message: "Successfully sent message!"})
					continue
				}
				//if len(pk.Servers) == 0
				s.SendMessage(pk.Message)
				_ = planet.WritePacket(&sunpacket.TextResponse{Type: sunpacket.TextResponseSuccess, Message: "Successfully sent message!"})
				continue
			}
		}
	}()
}
