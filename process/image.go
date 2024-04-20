package process

import (
	"image"
	"image/color"
	"math"
)

func vSplitImage(img image.Image, splitColor color.Color) []image.Image {
	rect := img.Bounds()
	parts := []image.Image{}
	lastY := rect.Min.Y
	inDivider := false
	totalDiff := float64(0)
	minLine := math.MaxFloat64
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		totalDiff = float64(0)
		for x := rect.Min.X; x < rect.Max.X; x++ {
			c := img.At(x, y)

			totalDiff += deltaE(c, splitColor)
		}
		avgDiff := totalDiff / float64(rect.Max.X-rect.Min.X)
		minLine = min(minLine, avgDiff)
		// spew.Dump("===", y, avgDiff)
		if avgDiff > 7_000 {
			inDivider = false
			continue
		}
		if inDivider {
			continue
		}
		part := crop(img, rect.Min.X, lastY, rect.Max.X, y)
		parts = append(parts, part)
		lastY = y
		inDivider = true
	}
	// spew.Dump("===", minLine)
	part := crop(img, rect.Min.X, lastY, rect.Max.X, rect.Max.Y)
	parts = append(parts, part)

	return parts
}

func hSplitImage(img image.Image, splitColor color.Color) []image.Image {
	rect := img.Bounds()

	dividerWidth := 0
	totalDiff := float64(0)
	minLine := math.MaxFloat64

	maxDivWidth := 0
	maxDivX := 0

	for x := rect.Min.X; x < rect.Max.X; x++ {
		totalDiff = float64(0)
		for y := rect.Min.Y + 3; y < rect.Max.Y-3; y++ {
			c := img.At(x, y)

			totalDiff += deltaE(c, splitColor)
		}
		avgDiff := totalDiff / float64(rect.Max.Y-rect.Min.Y)
		minLine = min(minLine, avgDiff)
		// spew.Dump("===", x, avgDiff)
		if avgDiff > 300 {
			if dividerWidth > 0 {
				if dividerWidth > maxDivWidth {
					maxDivWidth = dividerWidth
					maxDivX = x - dividerWidth/2
				}
				dividerWidth = 0
			}
		} else {
			dividerWidth++
		}
	}
	a := crop(img, rect.Min.X, rect.Min.Y, maxDivX, rect.Max.Y)
	b := crop(img, maxDivX, rect.Min.Y, rect.Max.X, rect.Max.Y)
	// parts = append(parts, part)

	return []image.Image{a, b}
}

func deltaE(color1, color2 color.Color) float64 {
	r1, g1, b1, _ := color1.RGBA()
	r2, g2, b2, _ := color2.RGBA()

	deltaR := float64(r1) - float64(r2)
	deltaG := float64(g1) - float64(g2)
	deltaB := float64(b1) - float64(b2)

	return math.Sqrt(deltaR*deltaR + deltaG*deltaG + deltaB*deltaB)
}
