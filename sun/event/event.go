/*
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

package event

// Credit to github.com/df-mc/dragonfly for the original code.

// Context represents the context of an event. Handlers of an event may call methods on the context to change
// the result of the event.
type Context struct {
	cancel bool
	after  []func(bool)
}

// C returns a new event context.
func C() *Context {
	return &Context{}
}

// Cancel cancels the context.
func (ctx *Context) Cancel() {
	ctx.cancel = true
}

// Continue calls the function f if the context is not cancelled. If it is cancelled, Continue will return
// immediately.
// These functions are not generally useful for handling events. See After() for executing code after the
// event happens.
func (ctx *Context) Continue(f func()) {
	if !ctx.cancel {
		f()
		for _, v := range ctx.after {
			v(ctx.cancel)
		}
	}
}

// Stop calls the function f if the context is cancelled. If it is not cancelled, Stop will return
// immediately.
// Stop does the opposite of Continue.
// These functions are not generally useful for handling events. See After() for executing code after the
// event happens.
func (ctx *Context) Stop(f func()) {
	if ctx.cancel {
		f()
		for _, v := range ctx.after {
			v(ctx.cancel)
		}
	}
}

// After calls the function passed after the action of the event has been completed, either by a call to
// (*Context).Continue() or (*Context).Stop().
// After can be executed multiple times to attach more functions to be called after the event is executed.
func (ctx *Context) After(f func(cancelled bool)) {
	ctx.after = append(ctx.after, f)
}
