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

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sunproxy/sun/sun/event"
	"sync"
)

type Ray struct {
	// conn is the connection to the client. e.g conn -> buffer, buffer -> remote
	conn *minecraft.Conn
	// remote is the remote server that this ray is targeting. e.g remote -> buffer, buffer -> conn
	remote *minecraft.Conn
	// bufferConn is the connection to the actual proxy server. e.g con -> buffer -> remote
	bufferConn   *minecraft.Conn
	Translations *TranslatorMappings
	transferring bool
	remoteMu     sync.Mutex
	TransferData struct {
		ScoreboardNames map[string]struct{}
	}
	handler   Handler
	handlerMu sync.RWMutex
}

type TranslatorMappings struct {
	OriginalEntityRuntimeID uint64
	OriginalEntityUniqueID  int64
	CurrentEntityRuntimeID  uint64
	CurrentEntityUniqueID   int64
}

func New(conn *minecraft.Conn) *Ray {
	return &Ray{conn: conn,
		TransferData: struct{ ScoreboardNames map[string]struct{} }{ScoreboardNames: make(map[string]struct{})},
		handler:      NopHandler{},
	}
}

// Remote /*
func (r *Ray) Remote() *minecraft.Conn {
	r.remoteMu.Lock()
	defer r.remoteMu.Unlock()
	return r.remote
}

// Handle Changes the rays handler...
func (r *Ray) Handle(handler Handler) {
	if r == nil {
		return
	}
	r.handlerMu.Lock()
	defer r.handlerMu.Unlock()
	if handler == nil {
		handler = NopHandler{}
	}
	r.handler = handler
}

// Handler Returns the current handler...
func (r *Ray) Handler() Handler {
	if r == nil {
		return NopHandler{}
	}
	r.handlerMu.RLock()
	handler := r.handler
	r.handlerMu.RUnlock()
	return handler
}

// Transferring /*
func (r *Ray) Transferring() bool {
	return r.transferring
}

func (r *Ray) Conn() *minecraft.Conn {
	return r.conn
}

func (r *Ray) SetTransferring(transferring bool) {
	r.transferring = transferring
}

// BufferConn /**
func (r *Ray) BufferConn() *minecraft.Conn {
	return r.bufferConn
}

func (r *Ray) SetBufferConn(rem *minecraft.Conn) {
	r.bufferConn = rem
}

func (r *Ray) SetRemote(rem *minecraft.Conn) {
	r.remoteMu.Lock()
	r.remote = rem
	r.remoteMu.Unlock()
}

func (r *Ray) HandleTransferDataSwap(bufferC *minecraft.Conn) {
	r.remoteMu.Lock()
	r.remote = bufferC
	r.bufferConn = nil
	r.updateTranslatorData(r.remote.GameData())
	r.remoteMu.Unlock()
}

type Handler interface {
	// HandlePacketReceive Called when the ray recieves a packet.
	HandlePacketReceive(ctx *event.Context, pk packet.Packet, ray *Ray)
	// HandlePacketSend Called right before or when a ray tries to forward a packet.
	HandlePacketSend(ctx *event.Context, pk packet.Packet, ray *Ray)
}

type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

func (n NopHandler) HandlePacketReceive(*event.Context, packet.Packet, *Ray) {}

func (n NopHandler) HandlePacketSend(*event.Context, packet.Packet, *Ray) {}
