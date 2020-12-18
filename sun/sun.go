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
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"log"
	"sync"
)

type Sun struct {
	Listener *minecraft.Listener
	Players  map[string]*Player
	Hub      IpAddr
	open     bool
}

func (s *Sun) main() {
	defer s.Listener.Close()
	for s.open {
		//Listener won't be closed unless it is manually done
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		pl := &Player{conn: conn.(*minecraft.Conn), Sun: s}
		rconn, err := minecraft.Dialer{
			ClientData:   pl.conn.ClientData(),
			IdentityData: pl.conn.IdentityData()}.Dial("raknet", s.Hub.ToString())
		if err != nil {
			fmt.Println(err)
			_ = s.Listener.Disconnect(conn.(*minecraft.Conn), text.Colourf("<red>You Have been Disconnected!</red>"))
			continue
		}
		pl.remote = &Remote{rconn, s.Hub}
		s.AddPlayer(pl)
	}
}

/**
Starts the server synchronously depending on the bool
*/
func (s *Sun) start(async bool) {
	if !s.open {
		s.open = true
	}
	if !async {
		s.main()
		return
	}
	go s.main()
}

/**
Starts the server synchronously
*/
func (s *Sun) Start() {
	s.start(false)
}

/**
Starts the server asynchronously
*/
func (s *Sun) StartAsync() {
	s.start(true)
}

/**
closes the Sun will cause the main() to break
*/
func (s *Sun) Close() {
	s.open = false
}

/*
Adds a player to the sun and readies them
*/
func (s *Sun) AddPlayer(player *Player) {
	s.Players[player.conn.IdentityData().Identity] = player
	//start the player up
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := player.conn.StartGame(player.remote.conn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := player.remote.conn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()
	s.handlePlayer(player)
}

/**
Closes a players session cleanly with a nice disconnection message!
*/
func (s *Sun) ClosePlayer(player *Player) {
	_ = s.Listener.Disconnect(player.conn, text.Colourf("<red>You Have been Disconnected!</red>"))
	_ = player.remote.conn.Close()
	delete(s.Players, player.conn.IdentityData().Identity)
}

func (s *Sun) handlePlayer(player *Player) {
	go func() {
		for {
			pk, err := player.conn.ReadPacket()
			if err != nil {
				s.ClosePlayer(player)
				return
			}
			err = player.remote.conn.WritePacket(pk)
			if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
				_ = s.Listener.Disconnect(player.conn, disconnect.Error())
				return
			}
		}
	}()
	go func() {
		for {
			pk, err := player.remote.conn.ReadPacket()
			if err != nil {
				s.ClosePlayer(player)
				return
			}
			if tpk, ok := pk.(*packet.Transfer); ok {
				s.TransferPlayer(player, IpAddr{Ip: tpk.Address, Port: tpk.Port})
				continue
			}
			err = player.conn.WritePacket(pk)
			if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
				_ = s.Listener.Disconnect(player.conn, disconnect.Error())
				return
			}
		}
	}()
}

/**
Changes a players remote and readies the connection
*/
func (s *Sun) TransferPlayer(player *Player, addr IpAddr) {
	//Dial the new server based on the ipaddr
	conn, err := minecraft.Dialer{
		ClientData:   player.conn.ClientData(),
		IdentityData: player.conn.IdentityData()}.Dial("raknet", addr.ToString())
	if err != nil {
		//cleanly close player
		s.ClosePlayer(player)
		return
	}
	//Start server
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	//send out Buffered packets now so they don't get sent to the wrong connection
	s.flushPlayer(player)
	player.remote = &Remote{conn, addr}
}

/**
Flushes both connections a player might have for transfer
*/
func (s *Sun) flushPlayer(player *Player) {
	err := player.conn.Flush()
	if err != nil {
		log.Println(err)
	}
	err = player.remote.conn.Flush()
	if err != nil {
		log.Println(err)
	}
}
