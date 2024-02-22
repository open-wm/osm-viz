package main

import (
	"image/color"
	"strconv"

	"github.com/fogleman/gg"
)

func drawRoad(dc *gg.Context, w *WayWithNodes, mapData *mapData) {
	lineWidth := 5.0
	lanes := w.Way.Tags.Find("lanes")
	if lanes != "" {
		laneFloat, err := strconv.ParseFloat(lanes, 64)
		if err != nil {
			return
		}
		lineWidth = laneFloat*5.0 + 1.0
	}
	dc.SetLineWidth(lineWidth)
	colors := map[string][3]float64{
		"motorway":          {251 / 255.0, 133.0 / 255.0, 0.0 / 255.0},   // Javier prado y evitamiento // orange
		"trunk":             {2 / 255.0, 48.0 / 255.0, 71.0 / 255.0},     // IDK we dont seem to have any
		"primary":           {251 / 255.0, 183.0 / 255.0, 0.0 / 255.0},   // Panamericana // yellow
		"secondary":         {102 / 255.0, 148.0 / 255.0, 171.0 / 255.0}, // la floresta, olguin // blue
		"tertiary":          {33 / 255.0, 157 / 255.0, 188 / 255.0},
		"unclassified":      {0.6, 0.6, 0.6},
		"residential":       {1.0, 1.0, 1.0},
		"service":           {0.6, 0.6, 0.6},
		"motorway_link":     {251 / 255.0, 183.0 / 255.0, 0.0 / 255.0}, // yellow
		"trunk_link":        {0.6, 0.6, 0.6},
		"primary_link":      {251 / 255.0, 183.0 / 255.0, 0.0 / 255.0},
		"secondary_link":    {2 / 255.0, 48.0 / 255.0, 71.0 / 255.0},
		"tertiary_link":     {0.6, 0.6, 0.6},
		"unclassified_link": {0.6, 0.6, 0.6},
		"residential_link":  {0.6, 0.6, 0.6},
		"service_link":      {0.6, 0.6, 0.6},
	}
	x0, y0 := processLatLon(w.Nodes[0], mapData)

	roadType := w.Way.Tags.Find("highway")
	// set colors according to their type
	if roadType == "" {
		return
	}

	if roadType == "residential" {
		dc.Push()
		dc.SetRGBA(0.6, 0.6, 0.6, 0.3)
		ratio := 0.8
		dc.MoveTo(x0-lineWidth*ratio, y0-lineWidth*ratio)
		for i := 1; i < len(w.Nodes); i++ {
			x1, y1 := processLatLon(w.Nodes[i], mapData)
			dc.LineTo(x1-lineWidth*ratio, y1-lineWidth*ratio)
		}
		dc.Stroke()
		dc.MoveTo(x0+lineWidth*ratio, y0+lineWidth*ratio)
		// vector to store the direction
		vector := [2]float64{}
		for i := 1; i < len(w.Nodes); i++ {
			x1, y1 := processLatLon(w.Nodes[i], mapData)
			dc.LineTo(x1+lineWidth*ratio, y1+lineWidth*ratio)

			if i == 1 {
				// vector to store the direction of the road
				vector = [2]float64{x1, y1}
			}
		}
		dc.Stroke()
		// draw a random rectangle to simulate a car
		if false {
			dc.MoveTo(x0, y0)
			x1 := x0 + (vector[0]-x0)*0.2
			y1 := y0 + (vector[1]-y0)*0.2
			dc.LineTo(x1, y1)
			dc.LineTo(x1, y1-10)
			dc.LineTo(x0, y0-10)
			dc.ClosePath()
			dc.SetRGBA(1, 0.6, 0.6, 1)
			dc.Fill()
			// end of drawing a car
			dc.Pop()
		}
	}

	x0, y0 = processLatLon(w.Nodes[0], mapData)

	dc.MoveTo(x0, y0)
	for i := 1; i < len(w.Nodes); i++ {
		x1, y1 := processLatLon(w.Nodes[i], mapData)
		// if y1 < canvasY/3 || y1 > canvasY*2/3 {
		// 	continue
		// }
		dc.LineTo(x1-2, y1-2)
	}
	c := colors[roadType]
	if c[0] == 0.0 && c[1] == 0.0 && c[2] == 0.0 {
		dc.SetRGBA(0.6, 0.6, 0.6, 0.3)
	} else {
		dc.SetColor(color.NRGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), 200})
	}
	dc.Stroke()

}
