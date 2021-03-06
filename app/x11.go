// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!android

package app

/*
Simple on-screen app debugging for X11. Not an officially supported
development target for apps, as screens with mice are very different
than screens with touch panels.

On Ubuntu 14.04 'Trusty', you may have to install these libraries:
sudo apt-get install libegl1-mesa-dev libgles2-mesa-dev libx11-dev
*/

/*
#cgo LDFLAGS: -lEGL -lGLESv2 -lX11

void runApp(void);
*/
import "C"
import (
	"runtime"
	"sync"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
)

var cb Callbacks

func run(callbacks Callbacks) {
	runtime.LockOSThread()
	cb = callbacks
	C.runApp()
}

//export onResize
func onResize(w, h int) {
	// TODO(nigeltao): don't assume 72 DPI. DisplayWidth / DisplayWidthMM
	// is probably the best place to start looking.
	geom.PixelsPerPt = 1
	geom.Width = geom.Pt(w)
	geom.Height = geom.Pt(h)
}

var touchEvents struct {
	sync.Mutex
	pending []event.Touch
}

func sendTouch(ty event.TouchType, x, y float32) {
	touchEvents.Lock()
	touchEvents.pending = append(touchEvents.pending, event.Touch{
		ID:   0,
		Type: ty,
		Loc: geom.Point{
			X: geom.Pt(x / geom.PixelsPerPt),
			Y: geom.Height - geom.Pt(y/geom.PixelsPerPt),
		},
	})
	touchEvents.Unlock()
}

//export onTouchStart
func onTouchStart(x, y float32) { sendTouch(event.TouchStart, x, y) }

//export onTouchMove
func onTouchMove(x, y float32) { sendTouch(event.TouchMove, x, y) }

//export onTouchEnd
func onTouchEnd(x, y float32) { sendTouch(event.TouchEnd, x, y) }

//export onDraw
func onDraw() {
	touchEvents.Lock()
	pending := touchEvents.pending
	touchEvents.pending = nil
	touchEvents.Unlock()
	if cb.Touch != nil {
		for _, e := range pending {
			cb.Touch(e)
		}
	}

	if cb.Draw != nil {
		cb.Draw()
	}
}

//export onStart
func onStart() {
	if cb.Start != nil {
		cb.Start()
	}
}

//export onStop
func onStop() {
	if cb.Stop != nil {
		cb.Stop()
	}
}
