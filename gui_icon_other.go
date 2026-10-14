//go:build !windows && (!linux || nox11)

package main

import "gioui.org/app"

func setWindowIcon(app.ViewEvent) {}
