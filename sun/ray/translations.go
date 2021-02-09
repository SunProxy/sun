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

package ray

/*
We store all the translation shit in here
because it looks bad in the sun.go
*/

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (r *Ray) TranslatePacket(pk packet.Packet) {
	switch pk := pk.(type) {
	case *packet.ActorEvent:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.ActorPickRequest:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
	case *packet.AddActor:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.AddItemActor:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.AddPainting:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.AddPlayer:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.AdventureSettings:
		pk.PlayerUniqueID = r.translateUniqueID(pk.PlayerUniqueID)
	case *packet.Animate:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.AnimateEntity:
		for i := range pk.EntityRuntimeIDs {
			pk.EntityRuntimeIDs[i] = r.translateRuntimeID(pk.EntityRuntimeIDs[i])
		}
	case *packet.BossEvent:
		pk.BossEntityUniqueID = r.translateUniqueID(pk.BossEntityUniqueID)
		pk.PlayerUniqueID = r.translateUniqueID(pk.PlayerUniqueID)
	case *packet.Camera:
		pk.CameraEntityUniqueID = r.translateUniqueID(pk.CameraEntityUniqueID)
		pk.TargetPlayerUniqueID = r.translateUniqueID(pk.TargetPlayerUniqueID)
	case *packet.CommandOutput:
		pk.CommandOrigin.PlayerUniqueID = r.translateUniqueID(pk.CommandOrigin.PlayerUniqueID)
	case *packet.CommandRequest:
		pk.CommandOrigin.PlayerUniqueID = r.translateUniqueID(pk.CommandOrigin.PlayerUniqueID)
	case *packet.ContainerOpen:
		pk.ContainerEntityUniqueID = r.translateUniqueID(pk.ContainerEntityUniqueID)
	case *packet.DebugInfo:
		pk.PlayerUniqueID = r.translateUniqueID(pk.PlayerUniqueID)
	case *packet.Emote:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.EmoteList:
		pk.PlayerRuntimeID = r.translateRuntimeID(pk.PlayerRuntimeID)
	case *packet.Event:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.Interact:
		pk.TargetEntityRuntimeID = r.translateRuntimeID(pk.TargetEntityRuntimeID)
	case *packet.MobArmourEquipment:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MobEffect:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MobEquipment:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MotionPredictionHints:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MoveActorAbsolute:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MoveActorDelta:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.MovePlayer:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
		pk.RiddenEntityRuntimeID = r.translateRuntimeID(pk.RiddenEntityRuntimeID)
	case *packet.NPCRequest:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.PlayerAction:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.PlayerList:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = r.translateUniqueID(pk.Entries[i].EntityUniqueID)
		}
	case *packet.RemoveActor:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
	case *packet.Respawn:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.SetActorData:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.SetActorLink:
		pk.EntityLink.RiddenEntityUniqueID = r.translateUniqueID(pk.EntityLink.RiddenEntityUniqueID)
		pk.EntityLink.RiderEntityUniqueID = r.translateUniqueID(pk.EntityLink.RiderEntityUniqueID)
	case *packet.SetActorMotion:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.SetLocalPlayerAsInitialised:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.SetScore:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = r.translateUniqueID(pk.Entries[i].EntityUniqueID)
		}
	case *packet.SetScoreboardIdentity:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = r.translateUniqueID(pk.Entries[i].EntityUniqueID)
		}
	case *packet.ShowCredits:
		pk.PlayerRuntimeID = r.translateRuntimeID(pk.PlayerRuntimeID)
	case *packet.SpawnParticleEffect:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
	case *packet.StartGame:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.StructureBlockUpdate:
		pk.Settings.LastEditingPlayerUniqueID = r.translateUniqueID(pk.Settings.LastEditingPlayerUniqueID)
	case *packet.StructureTemplateDataRequest:
		pk.Settings.LastEditingPlayerUniqueID = r.translateUniqueID(pk.Settings.LastEditingPlayerUniqueID)
	case *packet.TakeItemActor:
		pk.ItemEntityRuntimeID = r.translateRuntimeID(pk.ItemEntityRuntimeID)
		pk.TakerEntityRuntimeID = r.translateRuntimeID(pk.TakerEntityRuntimeID)
	case *packet.UpdateAttributes:
		pk.EntityRuntimeID = r.translateRuntimeID(pk.EntityRuntimeID)
	case *packet.UpdateEquip:
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
	case *packet.UpdatePlayerGameType:
		pk.PlayerUniqueID = r.translateUniqueID(pk.PlayerUniqueID)
	case *packet.UpdateTrade:
		pk.VillagerUniqueID = r.translateUniqueID(pk.VillagerUniqueID)
		pk.EntityUniqueID = r.translateUniqueID(pk.EntityUniqueID)
	}
}

func (r *Ray) translateRuntimeID(id uint64) uint64 {
	original := r.Translations.OriginalEntityRuntimeID
	current := r.Translations.CurrentEntityRuntimeID

	if original == id {
		return current
	} else if current == id {
		return original
	}
	return id
}

func (r *Ray) translateUniqueID(id int64) int64 {
	original := r.Translations.OriginalEntityUniqueID
	current := r.Translations.CurrentEntityUniqueID

	if original == id {
		return current
	} else if current == id {
		return original
	}
	return id
}

func (r *Ray) InitTranslators(data minecraft.GameData) {
	r.Translations = &TranslatorMappings{
		OriginalEntityRuntimeID: data.EntityRuntimeID,
		OriginalEntityUniqueID:  data.EntityUniqueID,
	}
	r.updateTranslatorData(data)
}

func (r *Ray) updateTranslatorData(data minecraft.GameData) {
	if r.Translations == nil {
		r.Translations = &TranslatorMappings{}
	}
	r.Translations.CurrentEntityRuntimeID = data.EntityRuntimeID
	r.Translations.CurrentEntityUniqueID = data.EntityUniqueID
}
