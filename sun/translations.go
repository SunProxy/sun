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

/*
We store all the translation shit in here
because it looks bad in the sun.go
*/

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

func TranslateClientEntityRuntimeIds(player *Player, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.CommandBlockUpdate:
		{
			pk.MinecartEntityRuntimeID = player.TranslateEntityRuntimeID(pk.MinecartEntityRuntimeID)
		}
	case *packet.Interact:
		{
			pk.TargetEntityRuntimeID = player.TranslateEntityRuntimeID(pk.TargetEntityRuntimeID)
		}
	case *packet.EmoteList:
	case *packet.Emote:
	case *packet.SetLocalPlayerAsInitialised:
	case *packet.NPCRequest:
	case *packet.PlayerAction:
	case *packet.MobEquipment:
		{
			pk.EntityRuntimeID = player.TranslateEntityRuntimeID(pk.EntityRuntimeID)
		}
	case *packet.MovePlayer:
		{
			pk.EntityRuntimeID = player.TranslateEntityRuntimeID(pk.EntityRuntimeID)
			pk.RiddenEntityRuntimeID = player.TranslateEntityRuntimeID(pk.RiddenEntityRuntimeID)
		}
	case *packet.ActorPickRequest:
		{
			pk.EntityUniqueID = player.TranslateEntityUniqueID(pk.EntityUniqueID)
		}
	}
}

func TranslateServerEntityRuntimeIds(player *Player, pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.UpdateTrade:
		{
			pk.EntityUniqueID = player.TranslateEntityUniqueID(pk.EntityUniqueID)
			pk.VillagerUniqueID = player.TranslateEntityUniqueID(pk.VillagerUniqueID)
		}
	case *packet.ShowCredits:
		{
			pk.PlayerRuntimeID = player.TranslateEntityRuntimeID(pk.PlayerRuntimeID)
		}
	case *packet.BossEvent:
		{
			pk.PlayerUniqueID = player.TranslateEntityUniqueID(pk.PlayerUniqueID)
			pk.BossEntityUniqueID = player.TranslateEntityUniqueID(pk.BossEntityUniqueID)
		}
	case *packet.Camera:
		{
			pk.CameraEntityUniqueID = player.TranslateEntityUniqueID(pk.CameraEntityUniqueID)
			pk.TargetPlayerUniqueID = player.TranslateEntityUniqueID(pk.TargetPlayerUniqueID)
		}
	case *packet.MotionPredictionHints:
	case *packet.Emote:
	case *packet.MoveActorDelta:
	case *packet.Event:
	case *packet.Respawn:
	case *packet.Animate:
	case *packet.SetActorMotion:
	case *packet.SetActorData:
	case *packet.MobArmourEquipment:
	case *packet.MobEquipment:
	case *packet.UpdateAttributes:
	case *packet.MobEffect:
	case *packet.ActorEvent:
	case *packet.MoveActorAbsolute:
		{
			pk.EntityRuntimeID = player.TranslateEntityRuntimeID(pk.EntityRuntimeID)
		}
	case *packet.MovePlayer:
		{
			pk.EntityRuntimeID = player.TranslateEntityRuntimeID(pk.EntityRuntimeID)
			pk.RiddenEntityRuntimeID = player.TranslateEntityRuntimeID(pk.RiddenEntityRuntimeID)
		}
	case *packet.DebugInfo:
	case *packet.UpdatePlayerGameType:
	case *packet.SpawnParticleEffect:
	case *packet.UpdateBlockSynced:
	case *packet.UpdateEquip:
	case *packet.RemoveActor:
		{
			pk.EntityUniqueID = player.TranslateEntityUniqueID(pk.EntityUniqueID)
		}
	case *packet.TakeItemActor:
		{
			pk.ItemEntityRuntimeID = player.TranslateEntityRuntimeID(pk.ItemEntityRuntimeID)
			pk.TakerEntityRuntimeID = player.TranslateEntityRuntimeID(pk.TakerEntityRuntimeID)
		}
	case *packet.AddItemActor:
	case *packet.AddActor:
	case *packet.AddPlayer:
		{
			pk.EntityRuntimeID = player.TranslateEntityRuntimeID(pk.EntityRuntimeID)
			pk.EntityUniqueID = player.TranslateEntityUniqueID(pk.EntityUniqueID)
		}
	}
}
