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
	"github.com/sunproxy/sun/sun/event"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"github.com/sunproxy/sun/sun/ray"
	"log"
	"runtime"
	"strings"
	"time"
)

func (s *Sun) handleRay(ray *ray.Ray) {
	go func() {
		for {
			pk, err := ray.Conn().ReadPacket()
			if err != nil {
				return
			}
			ray.TranslatePacket(pk)
			ctx := event.C()
			ctx.Continue(func() {
				switch pk := pk.(type) {
				case *packet.CommandRequest:
					args := strings.Split(pk.CommandLine, " ")
					switch args[0][1:] {
					case "status":
						if s.StatusCommand {
							err = ray.Conn().WritePacket(&packet.Text{
								Message:  text.Colourf("<yellow>---- <red>Status</red> ----</yellow>"),
								TextType: packet.TextTypeRaw,
							})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							stats := &runtime.MemStats{}
							runtime.ReadMemStats(stats)
							if err != nil {
								err = ray.Conn().WritePacket(&packet.Text{
									Message: text.Colourf("<red>Error retrieving " +
										"virtual memory statistics!</red>"),
									TextType: packet.TextTypeRaw,
								})
								if err != nil {
									_ = s.Logger.Debugf("Error Stopped Listener "+
										"Routine: %s", err.Error())
									return
								}
								return
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message: text.Colourf("<yellow>Total Ram Usage:</yellow>"+
									" <red>%v bytes</red>", stats.Alloc),
								TextType: packet.TextTypeRaw})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message: text.Colourf("<yellow>All Time Allocated Memory"+
									":</yellow> "+
									"<red>%v bytes</red>",
									stats.TotalAlloc),
								TextType: packet.TextTypeRaw})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message: text.Colourf("<yellow>Total GoRoutine Count:</yellow> <red>%v</red>",
									runtime.NumGoroutine()),
								TextType: packet.TextTypeRaw})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message: text.Colourf("<yellow>Total Player Count:<yellow> <red>%v</red>",
									len(s.Rays)),
								TextType: packet.TextTypeRaw})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message:  text.Colourf("<yellow>-----------------</yellow>"),
								TextType: packet.TextTypeRaw,
							})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							return
						}
					}
				case *packet.PlayerAction:
					if pk.ActionType == protocol.PlayerActionDimensionChangeDone && ray.Transferring() {
						ray.SetTransferring(false)

						old := ray.Remote()
						bufferC := ray.BufferConn()

						pos := bufferC.GameData().PlayerPosition
						err = ray.Conn().WritePacket(&packet.ChangeDimension{
							Dimension: packet.DimensionOverworld,
							Position:  pos,
						})
						if err != nil {
							s.BreakRay(ray)
						}
						_ = old.Close()
						ray.HandleTransferDataSwap(bufferC)
						log.Println("Successfully completed transfer for player", ray.Conn().IdentityData().DisplayName)
						return
					}
				}
				err = ray.Remote().WritePacket(pk)
				if err != nil {
					_ = s.Logger.Debugf("Error Stopped Listener "+
						"Routine: %s", err.Error())
					return
				}
			})
			ray.Handler().HandlePacketSend(ctx, pk, ray)
		}
	}()
	go func() {
		for {
			pk, err := ray.Remote().ReadPacket()
			if err != nil {
				_ = s.Logger.Debugf("Error Stopped Listener "+
					"Routine: %s", err.Error())
				return
			}
			ctx := event.C()
			ctx.Continue(func() {
				//Forward packet...
				err = ray.Conn().WritePacket(pk)
				if err != nil {
					_ = s.Logger.Debugf("Error Stopped Listener "+
						"Routine: %s", err.Error())
					return
				}
			})
			ray.Handler().HandlePacketReceive(ctx, pk, ray)
			ray.TranslatePacket(pk)
			//We won't allow plugins to override base listeners...
			switch pk := pk.(type) {
			case *packet.AvailableCommands:
				if s.StatusCommand {
					pk.Commands = append(pk.Commands, protocol.Command{
						Name:        "status",
						Description: "Provides information on the sun proxies load and player count!",
					})
					_ = ray.Conn().WritePacket(pk)
				}
			case *sunpacket.Transfer:
				err := s.TransferRay(ray, fmt.Sprintf("%s:%d", pk.Address, pk.Port))
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
				delete(ray.TransferData.ScoreboardNames, pk.ObjectiveName)
			case *packet.SetDisplayObjective:
				ray.TransferData.ScoreboardNames[pk.ObjectiveName] = struct{}{}
			}
		}
	}()
}

// TransferRay transfers a ray to a new server.
func (s *Sun) TransferRay(ray *ray.Ray, addr string) error {
	log.Println("Transfer request received for", ray.Conn().IdentityData().DisplayName)
	if ray.Transferring() {
		log.Printf("Transfer scrapped because a transfer request was already made for %s", ray.Conn().IdentityData().DisplayName)
		return fmt.Errorf("transfer scrapped because a transfer request was already made for %s", ray.Conn().IdentityData().DisplayName)
	}
	ray.SetTransferring(false)
	//Dial the new server based on the ipaddr
	idend := ray.Conn().IdentityData()
	//clear the xuid this might be the fix
	idend.XUID = ""
	conn, err := minecraft.Dialer{
		ClientData:   ray.Conn().ClientData(),
		IdentityData: idend}.Dial("raknet", addr)
	if err != nil {
		ray.SetTransferring(false)
		return err
	}
	ray.SetBufferConn(conn)
	//do spawn
	err = ray.BufferConn().DoSpawnTimeout(time.Minute)
	if err != nil {
		return err
	}
	for obj := range ray.TransferData.ScoreboardNames {
		err := ray.Conn().WritePacket(&packet.RemoveObjective{ObjectiveName: obj})
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
	err = ray.Conn().WritePacket(&packet.InventoryContent{
		WindowID: protocol.WindowIDInventory,
		Content:  items,
	})
	var gm int32
	//The gm should be the original gm of the remote player
	gm = ray.BufferConn().GameData().PlayerGameMode
	//if the gm is 5 we use the WorldGameMode as the players
	if gm == 5 {
		gm = ray.BufferConn().GameData().WorldGameMode
	}
	err = ray.Conn().WritePacket(&packet.SetPlayerGameType{
		GameType: gm,
	})
	if err != nil {
		return err
	}
	err = ray.Conn().WritePacket(&packet.ChangeDimension{
		Dimension: packet.DimensionNether,
		Position:  ray.Conn().GameData().PlayerPosition,
	})
	if err != nil {
		return err
	}
	//send empty chunk data.
	chunkX := int32(ray.Conn().GameData().PlayerPosition.X()) >> 4
	chunkZ := int32(ray.Conn().GameData().PlayerPosition.Z()) >> 4
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			err = ray.Conn().WritePacket(&packet.LevelChunk{
				Position:      [2]int32{chunkX + x, chunkZ + z},
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
