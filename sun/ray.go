package sun

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sunproxy/sun/sun/event"
	"github.com/sunproxy/sun/sun/ip_addr"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"github.com/sunproxy/sun/sun/ray"
	"github.com/sunproxy/sun/sun/remote"
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
					case "transfer":
						if s.TransferCommand {
							if len(args) > 1 {
								server := args[1]
								if ip, ok := s.Servers[server]; ok {
									err = ray.Conn().WritePacket(&packet.Text{
										Message: text.Colourf("<yellow>You Are Being Transferred To Server %s</yellow>",
											server),
										TextType: packet.TextTypeRaw})
									if err != nil {
										_ = s.Logger.Debugf("Error Stopped Listener "+
											"Routine: %s", err.Error())
										return
									}
									err := s.TransferRay(ray, ip)
									if err != nil {
										_ = ray.Conn().WritePacket(&packet.Text{
											Message:  text.Colourf("<red>An Occurred During Your Transfer Request!</red>"),
											TextType: packet.TextTypeRaw})
										return
									}
								} else {
									err = ray.Conn().WritePacket(&packet.Text{
										Message: text.Colourf("<red>Server %s Was Not Found In The Config.yml!</red>",
											server),
										TextType: packet.TextTypeRaw})
									if err != nil {
										_ = s.Logger.Debugf("Error Stopped Listener "+
											"Routine: %s", err.Error())
										return
									}
									return
								}
							}
							err = ray.Conn().WritePacket(&packet.Text{
								Message:  text.Colourf("<red>Please Provide a Server To Be Transferred To!</red>"),
								TextType: packet.TextTypeRaw})
							if err != nil {
								_ = s.Logger.Debugf("Error Stopped Listener "+
									"Routine: %s", err.Error())
								return
							}
							return
						}
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
					if pk.ActionType == packet.PlayerActionDimensionChangeDone && ray.Transferring() {
						ray.SetTransferring(false)

						old := ray.Remote().Conn
						bufferC := ray.BufferConn()

						pos := bufferC.Conn.GameData().PlayerPosition
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
				err = ray.Remote().Conn.WritePacket(pk)
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
			pk, err := ray.Remote().Conn.ReadPacket()
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
					_ = ray.Conn().WritePacket(pk)
					continue
				} else if s.StatusCommand {
					pk.Commands = append(pk.Commands, protocol.Command{
						Name:        "status",
						Description: "Provides information on the sun proxies load and player count!",
					})
					_ = ray.Conn().WritePacket(pk)
				}
			case *sunpacket.Transfer:
				err := s.TransferRay(ray, ip_addr.IpAddr{Address: pk.Address, Port: pk.Port})
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

/*
Changes a players remote and readies the connection
*/
func (s *Sun) TransferRay(ray *ray.Ray, addr ip_addr.IpAddr) error {
	log.Println("Transfer request received for", ray.Conn().IdentityData().DisplayName)
	if ray.Transferring() {
		log.Println("Transfer scrapped because it was already transferring for", ray.Conn().IdentityData().DisplayName)
		return fmt.Errorf("transfer scrapped because it was already transferring for %s", ray.Conn().IdentityData().DisplayName)
	}
	ray.SetTransferring(false)
	//Dial the new server based on the ipaddr
	idend := ray.Conn().IdentityData()
	//clear the xuid this might be the fix
	idend.XUID = ""
	conn, err := minecraft.Dialer{
		ClientData:   ray.Conn().ClientData(),
		IdentityData: idend}.Dial("raknet", addr.ToString())
	if err != nil {
		ray.SetTransferring(false)
		return err
	}
	ray.SetBufferConn(remote.New(conn, addr))
	//do spawn
	err = ray.BufferConn().Conn.DoSpawnTimeout(time.Minute)
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
	//Declare the gamemode variable
	var gamemode int32
	//The Gamemode should be the original gamemode of the remote player
	gamemode = ray.BufferConn().Conn.GameData().PlayerGameMode
	//if the gamemode 5 we use the WorldGameMode as the players
	if gamemode == 5 {
		gamemode = ray.BufferConn().Conn.GameData().WorldGameMode
	}
	err = ray.Conn().WritePacket(&packet.SetPlayerGameType{
		GameType: gamemode,
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
