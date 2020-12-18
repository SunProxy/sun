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

func (s *Sun) Start() {
	if !s.open {
		s.open = true
	}
	go s.main()
}

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
	s.flushPlayer(player)
	conn, err := minecraft.Dialer{
		ClientData:   player.conn.ClientData(),
		IdentityData: player.conn.IdentityData()}.Dial("raknet", addr.ToString())
	if err != nil {
		_ = s.Listener.Disconnect(player.conn, text.Colourf("<red>You Have been Disconnected!</red>"))
		return
	}
	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	player.remote = &Remote{conn, addr}
}

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
