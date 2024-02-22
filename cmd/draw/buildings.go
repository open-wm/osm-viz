package main

import (
	"image/color"
	"strconv"

	"github.com/fogleman/gg"
)

func drawBuilding(dc *gg.Context, w *WayWithNodes, mapData *mapData) {
	// Draw the base of the building

	// dc.SetLineWidth(1.0)
	// for i := 0; i < len(w.Nodes)-1; i++ {
	// 	x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
	// 	dc.LineTo(x1, y1)
	// }
	// dc.ClosePath()
	// dc.SetRGB(1, 0, 0)
	// dc.Stroke()

	// return
	// // draw height (disabled because it doesnt look good)

	height := 20.0
	levels := w.Way.Tags.Find("building:levels")
	dc.SetLineWidth(2.0)
	dc.SetRGBA(0.5, 0.5, 0.8, 0.3)
	if levels != "" {
		level, err := strconv.Atoi(levels)
		if err != nil {
		} else {
			height = float64(level) * 20.0 // make floors not as big
		}
		dc.SetRGBA(0.5, 0.5, 0.8, 1)
		dc.SetLineWidth(1.0)
	} else {
		// if it doesnt have the level info do?
		// return
	}
	var pairXY [][2]float64 = [][2]float64{}
	for i := 0; i < len(w.Nodes)-1; i++ {
		x1, y1 := processLatLon(w.Nodes[i], mapData)
		pairXY = append(pairXY, [2]float64{x1, y1 - height})

		dc.NewSubPath()
		dc.MoveTo(x1, y1)
		// for each node in the building lets draw a line that goes up
		dc.LineTo(x1, y1-height)
		x2, y2 := processLatLon(w.Nodes[i+1], mapData)
		dc.SetRGBA(0.5, 0.5, 0.8, 1)
		dc.SetLineWidth(0.5)
		dc.Stroke()
		pairXY = append(pairXY, [2]float64{x2, y2 - height})

		dc.Push()
		dc.SetRGB(1, 0, 0)

		start := min(y1, y2)
		end := max(y1, y2)
		// startX := min(x1, x2)
		// endX := max(x1, x2)

		linear := gg.NewLinearGradient(0, start-height, 0, end)
		// 119, 141, 169
		linear.AddColorStop(0, color.RGBA{180, 194.0, 213.0, 255})
		linear.AddColorStop(1, color.RGBA{220, 220, 220, 255})
		dc.SetFillStyle(linear)
		dc.NewSubPath()
		dc.MoveTo(x1, y1)
		dc.LineTo(x1, y1-height)
		dc.LineTo(x2, y2-height)
		dc.LineTo(x2, y2)
		dc.ClosePath()
		// dc.DrawRectangle(x1, end-height, endX-startX, height)
		dc.Fill()
		dc.Pop()
	}

	// draw the top of the building
	dc.SetLineWidth(1.0)
	dc.SetRGBA(0.5, 0.5, 0.8, 1)
	dc.SetRGBA(0.75, 0.75, 1, 1)
	dc.SetRGBA(0.90, 0.90, 0.90, 1)
	dc.MoveTo(pairXY[0][0], pairXY[0][1])
	for _, pair := range pairXY {
		dc.LineTo(pair[0], pair[1])
	}
	dc.ClosePath()
	dc.Fill()
}
