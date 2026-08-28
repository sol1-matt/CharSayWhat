//go:build !windows

package main

import (
	"fmt"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func buildOutput(input string) string {
	chars := analyzeText(input)
	if len(chars) == 0 {
		return ""
	}
	widths := make([]int, len(chars))
	for i, c := range chars {
		w := len(c.pos)
		if len(c.char) > w {
			w = len(c.char)
		}
		if len(c.typ) > w {
			w = len(c.typ)
		}
		widths[i] = w
	}
	pos := make([]string, len(chars))
	chs := make([]string, len(chars))
	typ := make([]string, len(chars))
	for i, c := range chars {
		pos[i] = c.pos
		chs[i] = c.char
		typ[i] = c.typ
	}
	const labelW = 5
	var sb strings.Builder
	writeRow := func(label string, vals []string) {
		fmt.Fprintf(&sb, "%-*s", labelW, label)
		for i, v := range vals {
			fmt.Fprintf(&sb, "  %-*s", widths[i], v)
		}
		sb.WriteByte('\n')
	}
	writeRow("Pos", pos)
	writeRow("Char", chs)
	writeRow("Type", typ)
	return sb.String()
}

func runGUI() {
	go func() {
		var w app.Window
		w.Option(app.Title("CharSayWhat"), app.Size(unit.Dp(900), unit.Dp(600)))
		if err := guiLoop(&w); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func guiLoop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var ops op.Ops
	var inputEd widget.Editor
	inputEd.SingleLine = false

	var outputEd widget.Editor
	outputEd.SingleLine = false

	var lastInput string

	for {
		e := w.Event()
		switch e := e.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			current := inputEd.Text()
			if current != lastInput {
				lastInput = current
				outputEd.SetText(buildOutput(current))
			}

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Editor(th, &inputEd, "Paste or type text here...").Layout(gtx)
					})
				}),
				layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th, &outputEd, "")
						ed.Font = font.Font{Typeface: "Go Mono"}
						ed.TextSize = unit.Sp(13)
						return ed.Layout(gtx)
					})
				}),
			)

			e.Frame(gtx.Ops)

		case app.DestroyEvent:
			return e.Err
		}
	}
}
