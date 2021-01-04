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
)

type Ray struct {
	conn         *minecraft.Conn
	remote       *Remote
	bufferConn   *Remote
	Translations *TranslatorMappings
	transferring bool
}

type TranslatorMappings struct {
	OriginalEntityRuntimeID uint64
	OriginalEntityUniqueID  int64
	CurrentEntityRuntimeID  uint64
	CurrentEntityUniqueID   int64
}

/*
Translates the entityUniqueID from a given packet to fix mix matched IDs
*/
func (r *Ray) TranslateEntityUniqueID(entityUniqueID int64) int64 {
	if entityUniqueID == r.Translations.OriginalEntityUniqueID {
		return r.Translations.CurrentEntityUniqueID
	} else if entityUniqueID == r.Translations.CurrentEntityUniqueID {
		return r.Translations.OriginalEntityUniqueID
	}
	return entityUniqueID
}

/*
Translates the entityRuntimeID from a given packet to fix mix matched IDs
*/
func (r *Ray) TranslateEntityRuntimeID(entityRuntimeID uint64) uint64 {
	if entityRuntimeID == r.Translations.OriginalEntityRuntimeID {
		return r.Translations.CurrentEntityRuntimeID
	} else if entityRuntimeID == r.Translations.CurrentEntityRuntimeID {
		return r.Translations.OriginalEntityRuntimeID
	}
	return entityRuntimeID
}

/*
Updates the TranslatorMappings for the said Player
*/
func (r *Ray) UpdateTranslations() {
	r.Translations.CurrentEntityRuntimeID = r.remote.conn.GameData().EntityRuntimeID
	r.Translations.CurrentEntityUniqueID = r.remote.conn.GameData().EntityUniqueID
}

/*
Should only be called when the player is first joined / added
*/
func (r *Ray) InitTranslations() {
	r.Translations = &TranslatorMappings{OriginalEntityUniqueID: r.remote.conn.GameData().EntityUniqueID,
		OriginalEntityRuntimeID: r.remote.conn.GameData().EntityRuntimeID}
	//safe as p.Translations is no longer nil and should still have the same data which is correct
	r.UpdateTranslations()
}

/**
Returns the Remote Connection the player has currently.
*/
func (r *Ray) Remote() *Remote {
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
