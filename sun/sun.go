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
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"go.uber.org/atomic"
	"log"
	"net"
	"sync"
)

type Sun struct {
	Listener  *minecraft.Listener
	Rays      map[string]*Ray
	Hub       IpAddr
	Planets   []*Planet
	PListener net.Listener
	Status    StatusProvider
}

type StatusProvider struct {
	status  *minecraft.ServerStatus
	playerc *atomic.Int64
}

func (s StatusProvider) ServerStatus(_ int, _ int) minecraft.ServerStatus {
	s.status.PlayerCount = int(s.playerc.Load())
	return *s.status
}

/*
Returns a new sun with config the specified config hence W
*/
func NewSunW(config Config) (*Sun, error) {
	status := StatusProvider{&config.Status, atomic.NewInt64(0)}
	listener, err := minecraft.ListenConfig{
		AuthenticationDisabled: true,
		StatusProvider:         status,
	}.Listen("raknet", fmt.Sprint(":", config.Port))
	if err != nil {
		return nil, err
	}
	//hehehehehehe
	plistener, err := net.Listen("tcp", ":42069")
	if err != nil {
		return nil, err
	}
	registerPackets()
	return &Sun{Listener: listener, PListener: plistener, Status: status, Rays: make(map[string]*Ray, config.Status.MaxPlayers), Hub: config.Hub, Planets: make([]*Planet, 0)}, nil
}

func registerPackets() {
	packet.Register(IDSunTransfer, func() packet.Packet { return &Transfer{} })
	packet.Register(IDSunText, func() packet.Packet { return &Text{} })
}

/*
Returns a new sun with a auto detected config
*/
func NewSun() (*Sun, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return NewSunW(cfg)
}

func (s *Sun) main() {
	defer s.Listener.Close()
	go func() {
		for {
			conn, err := s.PListener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			//TODO: Implement Ids for Planets
			pl := &Planet{conn: conn}
			s.AddPlanet(pl)
		}
	}()
	for {
		//Listener won't be closed unless it is manually done
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		ray := &Ray{conn: conn.(*minecraft.Conn)}
		rconn, err := minecraft.Dialer{
			ClientData:   ray.conn.ClientData(),
			IdentityData: ray.conn.IdentityData()}.Dial("raknet", s.Hub.ToString())
		if err != nil {
			log.Println(err)
			_ = s.Listener.Disconnect(conn.(*minecraft.Conn),
				text.Colourf("<red>You Have been Disconnected!</red>"))
			continue
		}
		ray.remote = &Remote{rconn, s.Hub}
		s.MakeRay(ray)
	}
}

/*
Starts the proxy.
*/
func (s *Sun) Start() {
	s.main()
}

/*
Adds a player to the sun and readies them
*/
func (s *Sun) MakeRay(ray *Ray) {
	s.Rays[ray.conn.IdentityData().Identity] = ray
	//start the player up
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := ray.conn.StartGame(ray.remote.conn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := ray.remote.conn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()
	//start translator
	ray.InitTranslations()
	//Add to player count
	s.Status.playerc.Add(1)
	//Start the two listener functions
	s.handleRay(ray)
}

/*
Closes a players session cleanly with a nice disconnection message!
*/
func (s *Sun) BreakRay(ray *Ray) {
	_ = s.Listener.Disconnect(ray.conn, text.Colourf("<red>You Have been Disconnected!</red>"))
	_ = ray.remote.conn.Close()
	delete(s.Rays, ray.conn.IdentityData().Identity)
}

func (s *Sun) handleRay(ray *Ray) {
	go func() {
		for {
			pk, err := ray.conn.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pk.(type) {
			case *packet.PlayerAction:
				//hehehehehehehehehehehehehehehehehehehehehehehe thx
				if pk.ActionType == packet.PlayerActionDimensionChangeDone && ray.Transferring() {
					ray.transferring = false

					old := ray.Remote().conn
					bufferC := ray.bufferConn

					pos := bufferC.conn.GameData().PlayerPosition
					_ = ray.conn.WritePacket(&packet.ChangeDimension{
						Dimension: packet.DimensionOverworld,
						Position:  pos,
					})

					_ = old.Close()

					ray.remote = bufferC
					ray.bufferConn = nil
				}
			}
			TranslateClientEntityRuntimeIds(ray, pk)
			err = ray.remote.conn.WritePacket(pk)
			if err != nil {
				return
			}
		}
	}()
	go func() {
		for {
			pk, err := ray.remote.conn.ReadPacket()
			if err != nil {
				return
			}
			TranslateServerEntityRuntimeIds(ray, pk)
			if pk, ok := pk.(*Transfer); ok {
				s.TransferRay(ray, IpAddr{Address: pk.Address, Port: pk.Port})
				return
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
			err = ray.conn.WritePacket(pk)
			if err != nil {
				return
			}
		}
	}()
}

func (s *Sun) SendMessageToServers(Message string, Servers []string) {
	for _, server := range Servers {
		for _, ray := range s.Rays {
			if ray.Remote().Addr().ToString() == server {
				_ = ray.conn.WritePacket(&packet.Text{Message: Message, TextType: packet.TextTypeRaw})
			}
		}
	}
}

/*
SendMessage is used for sending a Sun wide message to all the connected clients
*/
func (s *Sun) SendMessage(Message string) {
	for _, ray := range s.Rays {
		//Send raw chat to each player as client will accept it
		_ = ray.conn.WritePacket(&packet.Text{Message: Message, TextType: packet.TextTypeRaw})
	}
}

/*
Changes a players remote and readies the connection
*/
func (s *Sun) TransferRay(ray *Ray, addr IpAddr) {
	if ray.transferring {
		return
	}
	ray.transferring = true
	//Dial the new server based on the ipaddr
	conn, err := minecraft.Dialer{
		ClientData:   ray.conn.ClientData(),
		IdentityData: ray.conn.IdentityData()}.Dial("raknet", addr.ToString())
	if err != nil {
		//cleanly close player
		s.BreakRay(ray)
		return
	}
	//do spawn
	err = ray.bufferConn.conn.DoSpawn()
	if err != nil {
		//cleanly close player
		s.BreakRay(ray)
		return
	}
	//Another twisted copy because fuk im lazy
	ray.bufferConn = &Remote{conn, addr}
	_ = ray.conn.WritePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  ray.conn.GameData().PlayerPosition,
	})
	//send empty chunk data THX TWISTED IM LAZY lmao.......
	chunkX := int32(ray.conn.GameData().PlayerPosition.X()) >> 4
	chunkZ := int32(ray.conn.GameData().PlayerPosition.Z()) >> 4
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			_ = ray.conn.WritePacket(&packet.LevelChunk{
				ChunkX:        chunkX + x,
				ChunkZ:        chunkZ + z,
				SubChunkCount: 0,
				RawPayload:    make([]byte, 256),
			})
		}
	}
	s.handleRay(ray)
}

/*
Flushes both connections a player might have for transfer
*/
func (s *Sun) flushPlayer(ray *Ray) {
	err := ray.conn.Flush()
	if err != nil {
		log.Println(err)
	}
	err = ray.remote.conn.Flush()
	if err != nil {
		log.Println(err)
	}
}

func (s *Sun) handlePlanet(planet *Planet) {

}

func (s *Sun) AddPlanet(planet *Planet) {

}
