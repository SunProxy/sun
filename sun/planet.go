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
