//go:build linux && !nox11

package main

/*
#cgo pkg-config: x11
#include <X11/Xlib.h>
#include <X11/Xatom.h>

// Format-32 properties are passed to Xlib as arrays of long, whatever its size.
static void setNetWMIcon(void *dpy, unsigned long win, unsigned long *data, int n) {
	Display *d = dpy;
	Atom icon = XInternAtom(d, "_NET_WM_ICON", False);
	XChangeProperty(d, (Window)win, icon, XA_CARDINAL, 32, PropModeReplace, (unsigned char *)data, n);
	XFlush(d);
}
*/
import "C"

import "gioui.org/app"

// setWindowIcon sets the title-bar/taskbar icon on X11. Gio has no icon
// option, so this writes _NET_WM_ICON directly. Gio calls XInitThreads, so
// using its Display from the event goroutine is safe. Wayland has no
// per-window icon; there the compositor uses the .desktop file matching
// app.ID (see `./build.sh install`).
func setWindowIcon(e app.ViewEvent) {
	x, ok := e.(app.X11ViewEvent)
	if !ok || x.Display == nil || x.Window == 0 {
		return
	}
	icon := netWMIconData()
	if len(icon) == 0 {
		return
	}
	data := make([]C.ulong, len(icon))
	for i, v := range icon {
		data[i] = C.ulong(v)
	}
	C.setNetWMIcon(x.Display, C.ulong(x.Window), &data[0], C.int(len(data)))
}
