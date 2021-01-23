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
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"log"
	"runtime"
	"strings"
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
	TransferData struct {
		scoreboardNames map[string]struct{}
	}
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
			case *packet.CommandRequest:
				args := strings.Split(pk.CommandLine, " ")
				switch args[0][1:] {
				case "transfer":
					if s.TransferCommand {
						if len(args) > 1 {
							server := args[1]
							if ip, ok := s.Servers[server]; ok {
								err = ray.conn.WritePacket(&packet.Text{
									Message: text.Colourf("<yellow>You Are Being Transferred To Server %s</yellow>",
										server),
									TextType: packet.TextTypeRaw})
								if err != nil {
									return
								}
								err := s.TransferRay(ray, ip)
								if err != nil {
									_ = ray.conn.WritePacket(&packet.Text{
										Message:  text.Colourf("<red>An Occurred During Your Transfer Request!</red>"),
										TextType: packet.TextTypeRaw})
									continue
								}
							} else {
								err = ray.conn.WritePacket(&packet.Text{
									Message: text.Colourf("<red>Server %s Was Not Found In The Config.yml!</red>",
										server),
									TextType: packet.TextTypeRaw})
								if err != nil {
									return
								}
								continue
							}
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message:  text.Colourf("<red>Please Provide a Server To Be Transferred To!</red>"),
							TextType: packet.TextTypeRaw})
						if err != nil {
							return
						}
						continue
					}
				case "status":
					if s.StatusCommand {
						err = ray.conn.WritePacket(&packet.Text{
							Message:  text.Colourf("<yellow>---- <red>Status</red> ----</yellow>"),
							TextType: packet.TextTypeRaw,
						})
						if err != nil {
							return
						}
						stats := &runtime.MemStats{}
						runtime.ReadMemStats(stats)
						if err != nil {
							err = ray.conn.WritePacket(&packet.Text{
								Message: text.Colourf("<red>Error retrieving " +
									"virtual memory statistics!</red>"),
								TextType: packet.TextTypeRaw,
							})
							if err != nil {
								return
							}
							continue
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message: text.Colourf("<yellow>Total Ram Usage:</yellow>"+
								" <red>%v bytes</red>", stats.Alloc),
							TextType: packet.TextTypeRaw})
						if err != nil {
							return
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message: text.Colourf("<yellow>All Time Allocated Memory"+
								":</yellow> "+
								"<red>%v bytes</red>",
								stats.TotalAlloc),
							TextType: packet.TextTypeRaw})
						if err != nil {
							return
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message: text.Colourf("<yellow>Total GoRoutine Count:</yellow> <red>%v</red>",
								runtime.NumGoroutine()),
							TextType: packet.TextTypeRaw})
						if err != nil {
							return
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message: text.Colourf("<yellow>Total Player Count:<yellow> <red>%v</red>",
								len(s.Rays)),
							TextType: packet.TextTypeRaw})
						if err != nil {
							return
						}
						err = ray.conn.WritePacket(&packet.Text{
							Message:  text.Colourf("<yellow>-----------------</yellow>"),
							TextType: packet.TextTypeRaw,
						})
						if err != nil {

							return
						}
						continue
					}
				}
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
						s.BreakRay(ray)
					}
					_ = old.Close()
					ray.remoteMu.Lock()
					ray.remote = bufferC
					ray.bufferConn = nil
					ray.updateTranslatorData(ray.remote.conn.GameData())
					ray.remoteMu.Unlock()
					log.Println("Successfully completed transfer for player", ray.conn.IdentityData().DisplayName)
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
			switch pk := pk.(type) {
			case *packet.AvailableCommands:
				if s.TransferCommand {
					var servers []string
					var overloads []protocol.CommandOverload
					for name := range s.Servers {
						servers = append(servers, name)
					}
					overloads = append(overloads, protocol.CommandOverload{
						Parameters: []protocol.CommandParameter{
							{Name: "server",
								Type: protocol.CommandArgEnum | protocol.CommandArgValid,
								Enum: protocol.CommandEnum{
									Type:    "server",
									Options: servers,
								},
							},
						},
					})
					pk.Commands = append(pk.Commands, protocol.Command{
						Name:        "transfer",
						Description: "Transfer to another server!",
						Overloads:   overloads,
					})
					_ = ray.conn.WritePacket(pk)
					continue
				} else if s.StatusCommand {
					pk.Commands = append(pk.Commands, protocol.Command{
						Name:        "status",
						Description: "Provides information on the sun proxies load and player count!",
					})
					_ = ray.conn.WritePacket(pk)
				}
			case *sunpacket.Transfer:
				err := s.TransferRay(ray, IpAddr{Address: pk.Address, Port: pk.Port})
				if err != nil {
					log.Printf("An Occurred During A transfer request, Error: %s\n!", err.Error())
				}
				continue
			case *sunpacket.Text:
				//Only iterate if we have to.
				if len(pk.Servers) > 0 {
					//in a new routine because of the iteration
					go s.SendMessageToServers(pk.Message, pk.Servers)
					continue
				}
				//if len(pk.Servers) == 0
				s.SendMessage(pk.Message)
				continue
			case *packet.RemoveObjective:
				delete(ray.TransferData.scoreboardNames, pk.ObjectiveName)
			case *packet.SetDisplayObjective:
				ray.TransferData.scoreboardNames[pk.ObjectiveName] = struct{}{}
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
func (s *Sun) TransferRay(ray *Ray, addr IpAddr) error {
	log.Println("Transfer request received for", ray.conn.IdentityData().DisplayName)
	if ray.transferring {
		log.Println("Transfer scrapped because it was already transferring for", ray.conn.IdentityData().DisplayName)
		return fmt.Errorf("transfer scrapped because it was already transferring for %s", ray.conn.IdentityData().DisplayName)
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
		ray.transferring = false
		return err
	}
	ray.bufferConn = &Remote{conn: conn, addr: addr}
	//do spawn
	err = ray.BufferConn().conn.DoSpawnTimeout(time.Minute)
	if err != nil {
		return err
	}
	for obj := range ray.TransferData.scoreboardNames {
		err := ray.conn.WritePacket(&packet.RemoveObjective{ObjectiveName: obj})
		if err != nil {
			return err
		}
	}
	//Make the empty item slice.
	var items []protocol.ItemInstance
	for i := 0; i < 36; i++ {
		items = append(items, protocol.ItemInstance{
			StackNetworkID: 0,
			Stack:          protocol.ItemStack{},
		})
	}
	//Clear the player inventory.
	err = ray.conn.WritePacket(&packet.InventoryContent{
		WindowID: protocol.WindowIDInventory,
		Content:  items,
	})
	//Declare the gamemode variable
	var gamemode int32
	//The Gamemode should be the original gamemode of the remote player
	gamemode = ray.BufferConn().conn.GameData().PlayerGameMode
	//if the gamemode 5 we use the WorldGameMode as the players
	if gamemode == 5 {
		gamemode = ray.BufferConn().conn.GameData().WorldGameMode
	}
	err = ray.conn.WritePacket(&packet.SetPlayerGameType{
		GameType: gamemode,
	})
	if err != nil {
		return err
	}
	err = ray.conn.WritePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  ray.conn.GameData().PlayerPosition,
	})
	if err != nil {
		return err
	}
	//send empty chunk data.
	chunkX := int32(ray.conn.GameData().PlayerPosition.X()) >> 4
	chunkZ := int32(ray.conn.GameData().PlayerPosition.Z()) >> 4
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			err = ray.conn.WritePacket(&packet.LevelChunk{
				ChunkX:        chunkX + x,
				ChunkZ:        chunkZ + z,
				SubChunkCount: 0,
				RawPayload:    emptychunk,
			})
			if err != nil {
				return err
			}
		}
	}
	return err
}
