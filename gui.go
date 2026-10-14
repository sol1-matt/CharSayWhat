//go:build !windows

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// ── palette ─────────────────────────────────────────────────────────────────
// Mirrors the CSS in gui_windows.go's htmlPage; keep the two in step.

func rgb(c uint32) color.NRGBA {
	return color.NRGBA{R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c), A: 0xff}
}

var (
	colBg      = rgb(0x1e1e1e)
	colFg      = rgb(0xd4d4d4)
	colInputBg = rgb(0x252526)
	colAccent  = rgb(0x007acc)
	colHint    = rgb(0x5a5a5a)
	colLabel   = rgb(0x666666)
	colBorder  = rgb(0x2d2d2d)
)

var typeColors = map[string]color.NRGBA{
	"lowercase":   rgb(0x9cdcfe),
	"uppercase":   rgb(0xc586c0),
	"digit":       rgb(0xb5cea8),
	"space":       rgb(0x484848),
	"newline":     rgb(0x484848),
	"tab":         rgb(0x484848),
	"punctuation": rgb(0xce9178),
	"symbol":      rgb(0x4fc1ff),
	"other":       rgb(0xd4d4d4),
}

var monoFont = font.Font{Typeface: "Go Mono"}

// Grid geometry, matching the WebView2 table: a 52px label column and fixed
// heights for the Pos and Char rows. Character columns take their natural
// width (at least cellW) when the window has room; otherwise they shrink
// evenly towards the width of their Pos/Char text and the Type row wraps and
// grows taller, the way the browser's automatic table layout does. Only when
// every column is at that minimum does the grid scroll sideways.
const (
	labelW   = unit.Dp(52)
	cellW    = unit.Dp(56)
	cellPadX = unit.Dp(3) // per side
	cellPadY = unit.Dp(4) // per side
	posH     = unit.Dp(24)
	charH    = unit.Dp(42)
	typeH    = unit.Dp(26) // minimum; grows to fit wrapped type names
	gridGap  = unit.Dp(14) // space under the grid for the scrollbar
)

// ── gui ─────────────────────────────────────────────────────────────────────

func runGUI() {
	go func() {
		var w app.Window
		w.Option(app.Title("CharSayWhat"), app.Size(unit.Dp(900), unit.Dp(600)))
		if err := guiLoop(&w); err != nil {
			fmt.Fprintln(os.Stderr, "charsaywhat: GUI failed:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func guiLoop(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette = material.Palette{Bg: colBg, Fg: colFg, ContrastBg: colAccent, ContrastFg: colFg}
	th.Face = monoFont.Typeface

	var ops op.Ops
	var inputEd widget.Editor
	grid := widget.List{List: layout.List{Axis: layout.Horizontal}}

	var lastInput string
	var chars []charInfo

	for {
		switch e := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			if current := inputEd.Text(); current != lastInput {
				lastInput = current
				chars = analyzeText(current)
			}

			paint.Fill(gtx.Ops, colBg)
			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					return inputPane(gtx, th, &inputEd)
				}),
				layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
					return outputPane(gtx, th, &grid, chars)
				}),
			)

			e.Frame(gtx.Ops)

		case app.ViewEvent:
			setWindowIcon(e)

		case app.DestroyEvent:
			return e.Err
		}
	}
}

// inputPane is the text box on a slightly lighter background, with the blue
// rule along its bottom edge.
func inputPane(gtx layout.Context, th *material.Theme, ed *widget.Editor) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			fillRect(gtx, colInputBg, image.Rectangle{Max: gtx.Constraints.Max})
			return layout.Inset{Top: 12, Bottom: 12, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = gtx.Constraints.Max
				es := material.Editor(th, ed, "Paste or type text here...")
				es.Font = monoFont
				es.TextSize = unit.Sp(14)
				es.Color = colFg
				es.HintColor = colHint
				es.SelectionColor = color.NRGBA{R: 0x26, G: 0x4f, B: 0x78, A: 0xff}
				return es.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(2))
			fillRect(gtx, colAccent, image.Rectangle{Max: size})
			return layout.Dimensions{Size: size}
		}),
	)
}

// outputPane lays the analysis out as a table: a label column, then one
// column per character, fitted to the window width where possible.
func outputPane(gtx layout.Context, th *material.Theme, grid *widget.List, chars []charInfo) layout.Dimensions {
	return layout.Inset{Top: 12, Bottom: 12, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if len(chars) == 0 {
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}

		cols := make([]column, len(chars))
		for i, c := range chars {
			cols[i] = newColumn(th, c)
		}
		g := fitGrid(gtx, cols, gtx.Constraints.Max.X-gtx.Dp(labelW))

		// Only as tall as the grid plus its scrollbar, so the bar sits under it.
		if h := g.posH + g.charH + g.typeH + gtx.Dp(gridGap); h < gtx.Constraints.Max.Y {
			gtx.Constraints.Max.Y = h
			gtx.Constraints.Min.Y = 0
		}
		return material.List(th, grid).Layout(gtx, len(cols)+1, func(gtx layout.Context, i int) layout.Dimensions {
			// The scrollbar is drawn over the list, so leave room under the grid.
			return layout.Inset{Bottom: gridGap}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if i == 0 {
					return labelColumn(gtx, th, g)
				}
				return cols[i-1].layout(gtx, g, i-1, i == len(cols))
			})
		})
	})
}

// column holds the three cell widgets for one character.
type column struct {
	pos, chr, typ layout.Widget
}

func newColumn(th *material.Theme, c charInfo) column {
	cellText := func(s string, size unit.Sp, col color.NRGBA, wrap bool) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			l := material.Label(th, size, s)
			l.Color = col
			l.Font = monoFont
			l.Alignment = text.Middle
			if wrap {
				l.WrapPolicy = text.WrapGraphemes
				l.LineHeight = unit.Sp(14)
			} else {
				l.MaxLines = 1
			}
			return l.Layout(gtx)
		}
	}
	typeColor, ok := typeColors[c.typ]
	if !ok {
		typeColor = colFg
	}
	return column{
		pos: cellText(c.pos, 11, colLabel, false),
		chr: cellText(c.char, 22, colFg, false),
		typ: cellText(c.typ, 11, typeColor, true),
	}
}

func (c column) layout(gtx layout.Context, g gridGeom, i int, last bool) layout.Dimensions {
	w := g.widths[i]
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return gridCell(gtx, w, g.posH, last, false, layout.Center, c.pos)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return gridCell(gtx, w, g.charH, last, false, layout.Center, c.chr)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return gridCell(gtx, w, g.typeH, last, true, layout.N, c.typ)
		}),
	)
}

// gridGeom is the grid's pixel geometry for one frame.
type gridGeom struct {
	widths             []int
	posH, charH, typeH int
}

// fitGrid sizes each column for the available width and works out how tall
// the Type row must be once names wrap into the narrower columns.
func fitGrid(gtx layout.Context, cols []column, avail int) gridGeom {
	padX, padY := 2*gtx.Dp(cellPadX), 2*gtx.Dp(cellPadY)
	pref := make([]int, len(cols))
	mins := make([]int, len(cols))
	for i, c := range cols {
		fixed := imax(measure(gtx, c.pos, 0).X, measure(gtx, c.chr, 0).X) + padX
		mins[i] = fixed
		pref[i] = imax(gtx.Dp(cellW), imax(fixed, measure(gtx, c.typ, 0).X+padX))
	}
	g := gridGeom{
		widths: fitWidths(pref, mins, avail),
		posH:   gtx.Dp(posH),
		charH:  gtx.Dp(charH),
		typeH:  gtx.Dp(typeH),
	}
	for i, c := range cols {
		g.typeH = imax(g.typeH, measure(gtx, c.typ, g.widths[i]-padX).Y+padY)
	}
	return g
}

// fitWidths shrinks columns to fit avail pixels, capping the widest first
// (water-filling) but never going below each column's minimum. If even the
// minimums don't fit, it returns them and the grid scrolls.
func fitWidths(pref, mins []int, avail int) []int {
	widthsAt := func(limit int) ([]int, int) {
		ws := make([]int, len(pref))
		total := 0
		for i := range pref {
			ws[i] = imax(mins[i], min(pref[i], limit))
			total += ws[i]
		}
		return ws, total
	}
	hi := 0
	for _, p := range pref {
		hi = imax(hi, p)
	}
	if ws, total := widthsAt(hi); total <= avail {
		return ws
	}
	// Largest per-column limit whose total still fits.
	lo := 0
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if _, total := widthsAt(mid); total <= avail {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	ws, _ := widthsAt(lo)
	return ws
}

func labelColumn(gtx layout.Context, th *material.Theme, g gridGeom) layout.Dimensions {
	row := func(label string, h int) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Dp(labelW), h)
			gtx.Constraints = layout.Exact(size)
			layout.Inset{Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min = image.Point{} // E keeps Min.Y; drop it to centre vertically
					l := material.Label(th, unit.Sp(11), label)
					l.Color = colLabel
					l.Font = monoFont
					return l.Layout(gtx)
				})
			})
			return layout.Dimensions{Size: size}
		})
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		row("Pos", g.posH),
		row("Char", g.charH),
		row("Type", g.typeH),
	)
}

// measure returns a widget's size without drawing it. maxW of 0 means
// unconstrained, giving the widget's natural single-line width.
func measure(gtx layout.Context, w layout.Widget, maxW int) image.Point {
	if maxW <= 0 {
		maxW = 1 << 20
	}
	m := op.Record(gtx.Ops)
	gtx.Constraints = layout.Constraints{Max: image.Pt(maxW, 1<<20)}
	size := w(gtx).Size
	m.Stop()
	return size
}

// gridCell draws one bordered cell. Each cell paints its own top and left
// border; the last column and bottom row also paint right and bottom, so
// shared edges are drawn once, like CSS border-collapse.
func gridCell(gtx layout.Context, width, height int, lastCol, lastRow bool, align layout.Direction, content layout.Widget) layout.Dimensions {
	size := image.Pt(width, height)
	gtx.Constraints = layout.Exact(size)

	px := gtx.Dp(1)
	inner := image.Rectangle{Min: image.Pt(px, px), Max: size}
	if lastCol {
		inner.Max.X -= px
	}
	if lastRow {
		inner.Max.Y -= px
	}
	fillRect(gtx, colBorder, image.Rectangle{Max: size})
	fillRect(gtx, colBg, inner)

	layout.Inset{Top: cellPadY, Bottom: cellPadY, Left: cellPadX, Right: cellPadX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return align.Layout(gtx, content)
	})
	return layout.Dimensions{Size: size}
}

func fillRect(gtx layout.Context, c color.NRGBA, r image.Rectangle) {
	paint.FillShape(gtx.Ops, c, clip.Rect(r).Op())
}
