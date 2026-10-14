//go:build !windows

package main

import (
	"embed"
	"image/color"
	"image/png"
	"path"
)

// Window icons, generated from charsaywhat.png by `./build.sh icons`.
//
//go:embed icons/*.png
var iconFS embed.FS

// netWMIconData decodes the embedded icons into the X11 _NET_WM_ICON layout:
// for each size, width and height followed by width*height non-premultiplied
// ARGB pixels, row by row.
func netWMIconData() []uint32 {
	entries, err := iconFS.ReadDir("icons")
	if err != nil {
		return nil
	}
	var out []uint32
	for _, e := range entries {
		f, err := iconFS.Open(path.Join("icons", e.Name()))
		if err != nil {
			continue
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		b := img.Bounds()
		out = append(out, uint32(b.Dx()), uint32(b.Dy()))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				out = append(out, uint32(c.A)<<24|uint32(c.R)<<16|uint32(c.G)<<8|uint32(c.B))
			}
		}
	}
	return out
}
