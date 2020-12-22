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
	"log"
	"sync"
)

type Sun struct {
	Listener *minecraft.Listener
	Rays     map[string]*Ray
	Hub      IpAddr
	open     bool
}

type StatusProvider struct {
	status *minecraft.ServerStatus
}

func (s StatusProvider) ServerStatus(playerCount int, _ int) minecraft.ServerStatus {
	s.status.PlayerCount = playerCount
	return *s.status
}

/*
Returns a new sun with config the specified config hence W
*/
func NewSunW(config Config) (*Sun, error) {
	listener, err := minecraft.ListenConfig{
		AuthenticationDisabled: true,
		StatusProvider:         StatusProvider{&config.Status},
	}.Listen("raknet", fmt.Sprint(":", config.Port))
	if err != nil {
		return nil, err
	}
	registerPackets()
	return &Sun{Listener: listener, Rays: make(map[string]*Ray, config.Status.MaxPlayers), Hub: config.Hub}, nil
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
	for s.open {
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
closes the Sun will cause the main() to break
*/
func (s *Sun) Close() {
	s.open = false
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
	//Dial the new server based on the ipaddr
	conn, err := minecraft.Dialer{
		ClientData:   ray.conn.ClientData(),
		IdentityData: ray.conn.IdentityData()}.Dial("raknet", addr.ToString())
	if err != nil {
		//cleanly close player
		s.BreakRay(ray)
		return
	}
	if ray.remote.conn != nil {
		_ = ray.remote.conn.Close()
	}
	ray.remote = &Remote{conn, addr}
	//Start server
	if err := ray.remote.conn.DoSpawn(); err != nil {
		panic(err)
	}
	//force dimension change
	_ = ray.conn.WritePacket(&packet.ChangeDimension{Dimension: packet.DimensionEnd, Position: conn.GameData().PlayerPosition})
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
