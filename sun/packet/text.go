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

package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

/*
Text is sent by the server to send a message to all the connected players on the proxy.
*/
type Text struct {
	/*
		Servers is an array of strings that contains the servers IP addresses the text message should be broadcast to
	*/
	Servers []string

	/*
		The text message
	*/
	Message string
}

func (pk *Text) ID() uint32 {
	return IDPlanetText
}

func (pk *Text) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Servers))
	w.Varuint32(&l)
	for _, v := range pk.Servers {
		w.String(&v)
	}
}

func (pk *Text) Unmarshal(r *protocol.Reader) {
	//if count == 0 we send it to all the connect clients.
	var count uint32
	r.Varuint32(&count)
	pk.Servers = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Servers[i])
	}
}

const (
	TextResponseSuccess = iota
	TextResponseBadRequest
	TextResponseTargetsNotFound
)

/*
TextResponse is sent in response to a tcp api call to the Text packet.
*/
type TextResponse struct {
	/*
		Type is const shown above that represents the status of the request.
	*/
	Type byte

	/*
		Message is a accompanied message to the status can be anything basically.
	*/
	Message string
}

func (t TextResponse) ID() uint32 {
	return IDPlanetTextResponse
}

func (t TextResponse) Marshal(w *protocol.Writer) {
	w.Uint8(&t.Type)
	w.String(&t.Message)
}

func (t TextResponse) Unmarshal(r *protocol.Reader) {
	r.Uint8(&t.Type)
	r.String(&t.Message)
}
