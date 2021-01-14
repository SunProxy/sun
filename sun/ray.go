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
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"log"
	"sync"
	"time"
)

type Ray struct {
	conn         *minecraft.Conn
	remote       *Remote
	bufferConn   *Remote
	Translations *TranslatorMappings
	transferring bool
	remoteMu     sync.Mutex
}

type TranslatorMappings struct {
	OriginalEntityRuntimeID uint64
	OriginalEntityUniqueID  int64
	CurrentEntityRuntimeID  uint64
	CurrentEntityUniqueID   int64
}

/**
Returns the Remote Connection the player has currently.
*/
func (r *Ray) Remote() *Remote {
	r.remoteMu.Lock()
	defer r.remoteMu.Unlock()
	return r.remote
}

/**
Returns a bool representing if a player is Transferring.
*/
func (r *Ray) Transferring() bool {
	return r.transferring
}

/**
BufferConn is the connection used to temp out new conns also named temp conn
*/
func (r *Ray) BufferConn() *Remote {
	return r.bufferConn
}

func (s *Sun) handleRay(ray *Ray) {
	go func() {
		for {
			pk, err := ray.conn.ReadPacket()
			if err != nil {
				return
			}
			ray.translatePacket(pk)
			switch pk := pk.(type) {
			case *packet.PlayerAction:
				if pk.ActionType == packet.PlayerActionDimensionChangeDone && ray.Transferring() {
					ray.transferring = false

					old := ray.Remote().conn
					bufferC := ray.bufferConn

					pos := bufferC.conn.GameData().PlayerPosition
					err = ray.conn.WritePacket(&packet.ChangeDimension{
						Dimension: packet.DimensionOverworld,
						Position:  pos,
					})
					if err != nil {
						continue
					}
					_ = old.Close()
					ray.remoteMu.Lock()
					ray.remote = bufferC
					ray.bufferConn = nil
					ray.updateTranslatorData(ray.remote.conn.GameData())
					ray.remoteMu.Unlock()
					log.Println("Successfully completed transfer for player ", ray.conn.IdentityData().DisplayName)
					continue
				}
			}
			err = ray.Remote().conn.WritePacket(pk)
			if err != nil {
				return
			}
		}
	}()
	go func() {
		for {
			pk, err := ray.Remote().conn.ReadPacket()
			if err != nil {
				continue
			}
			ray.translatePacket(pk)
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

/*
Changes a players remote and readies the connection
*/
func (s *Sun) TransferRay(ray *Ray, addr IpAddr) {
	log.Println("Transfer request received for ", ray.conn.IdentityData().DisplayName)
	if ray.transferring {
		log.Println("Transfer scrapped because it was already transferring for", ray.conn.IdentityData().DisplayName)
		return
	}
	ray.transferring = true
	//Dial the new server based on the ipaddr
	idend := ray.conn.IdentityData()
	//clear the xuid this might be the fix
	idend.XUID = ""
	conn, err := minecraft.Dialer{
		ClientData:   ray.conn.ClientData(),
		IdentityData: idend}.Dial("raknet", addr.ToString())
	if err != nil {
		log.Println("error dialing new server for transfer request for", ray.conn.IdentityData().DisplayName+"\n", err)
		ray.transferring = false
		return
	}
	ray.bufferConn = &Remote{conn: conn, addr: addr}
	//do spawn
	err = ray.BufferConn().conn.DoSpawnTimeout(time.Minute)
	if err != nil {
		//cleanly close player
		s.BreakRay(ray)
		return
	}
	err = ray.conn.WritePacket(&packet.SetScoreboardIdentity{
		ActionType: packet.ScoreboardIdentityActionClear,
		Entries:    nil,
	})
	if err != nil {
		log.Println("error clearing scoreboard for player", ray.conn.IdentityData().DisplayName+"\n", err)
		s.BreakRay(ray)
		return
	}
	err = ray.conn.WritePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  ray.conn.GameData().PlayerPosition,
	})
	if err != nil {
		log.Println("error sending the dimension change request to the player", ray.conn.IdentityData().DisplayName+"\n", err)
		s.BreakRay(ray)
		return
	}
	//Update Chunk Radius for players.
	_ = ray.conn.WritePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(ray.conn.GameData().PlayerPosition.X()),
			int32(ray.conn.GameData().PlayerPosition.Y()),
			int32(ray.conn.GameData().PlayerPosition.Z())},
		Radius: 12 >> 4,
	})
	//send empty chunk data.
	chunkX := int32(ray.conn.GameData().PlayerPosition.X()) >> 4
	chunkZ := int32(ray.conn.GameData().PlayerPosition.Z()) >> 4
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			_ = ray.conn.WritePacket(&packet.LevelChunk{
				ChunkX:        chunkX + x,
				ChunkZ:        chunkZ + z,
				SubChunkCount: 0,
				RawPayload:    emptychunk,
			})
		}
	}
}
