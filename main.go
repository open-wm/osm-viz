package main

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/fogleman/gg"
	"github.com/paulmach/osm"
)

type WayWithNodes struct {
	Way   *osm.Way
	Nodes []*osm.Node
}

func main() {
	file, err := os.Open("map.osm")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the file using paulmach/osm
	bytes, _ := io.ReadAll(file)
	mappy := &osm.OSM{}
	xml.Unmarshal(bytes, mappy)

	// Print the number of nodes

	canvasX := 5000.0
	canvasY := 5000.0
	dc := gg.NewContext(int(canvasX), int(canvasY))

	centroidLat := 0.0
	centroidLon := 0.0
	count := 0
	minLat := 1000.0
	maxLat := -1000.0
	minLon := 1000.0
	maxLon := -1000.0

	// ways := []*osm.Ways{}
	ways := []*WayWithNodes{}
	for _, w := range mappy.Ways {
		nway := WayWithNodes{Way: w}
		for _, n := range w.Nodes {
			for _, node := range mappy.Nodes {
				if n.ID == node.ID {
					nway.Nodes = append(nway.Nodes, node)
					n.Lat = node.Lat
					n.Lon = node.Lon
					if n.Lat < minLat {
						minLat = n.Lat
					}
					if n.Lat > maxLat {
						maxLat = n.Lat
					}
					if n.Lon < minLon {
						minLon = n.Lon
					}
					if n.Lon > maxLon {
						maxLon = n.Lon
					}
				}
			}
			centroidLat += n.Lat
			centroidLon += n.Lon
			count++
		}
		ways = append(ways, &nway)
	}
	centroidLat /= float64(count)
	centroidLon /= float64(count)
	println(centroidLat, centroidLon, minLat, maxLat, minLon, maxLon)
	zoom := 6.0
	factorX := canvasX / (maxLat - minLat) * zoom
	factorY := canvasY / (maxLat - minLat) * zoom
	println(factorX, factorY)

	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, canvasX, canvasY)
	dc.Fill()
	for _, w := range ways {
		dc.NewSubPath()
		if w.Way.Tags.AnyInteresting() {
			if w.Way.Tags.Find("highway") != "" {
				drawRoad(dc, w, centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
			}
			if w.Way.Tags.Find("building") != "" {
				drawBuilding(dc, w, centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
			}
		} else {
			continue
		}
	}
	dc.Fill()
	timestamp := time.Now().UnixMicro()
	dc.SavePNG(fmt.Sprintf("map-%d.png", timestamp))

}
func processLatLon(node *osm.Node, centroidLat float64, centroidLon float64, factorX float64, factorY float64, canvasX float64, canvasY float64) (float64, float64) {
	x0 := (node.Lon-centroidLon)*factorX + canvasX/2
	lambda := 300.0
	skew := (1.0 - x0/(canvasY/2.0)) * lambda
	y0 := (node.Lat-centroidLat)*-1.0*factorY*0.5 + canvasY/2 + skew
	return x0, y0
}

func drawRoad(dc *gg.Context, w *WayWithNodes, centroidLat float64, centroidLon float64, factorX float64, factorY float64, canvasX float64, canvasY float64) {
	// for _, t := range w.Way.Tags {
	// 	println(t.Key, t.Value)
	// }
	lineWidth := 5.0
	lanes := w.Way.Tags.Find("lanes")
	if lanes != "" {
		laneFloat, err := strconv.ParseFloat(lanes, 64)
		if err != nil {
			return
		}
		lineWidth = laneFloat*5.0 + 5.0
	}
	dc.SetLineWidth(lineWidth)
	colors := map[string][3]float64{
		"motorway":          {0.8, 0.8, 0.0}, // carretera
		"trunk":             {0.0, 0.5, 0.5}, // avenida gandes
		"primary":           {0.8, 0.5, 1.0}, // avenidas
		"secondary":         {0.6, 0.6, 0.6},
		"tertiary":          {0.6, 0.6, 0.6},
		"unclassified":      {0.6, 0.6, 0.6},
		"residential":       {0.6, 0.6, 0.6},
		"service":           {0.6, 0.6, 0.6},
		"motorway_link":     {0.6, 0.6, 0.6},
		"trunk_link":        {0.6, 0.6, 0.6},
		"primary_link":      {0.6, 0.6, 0.6},
		"secondary_link":    {0.6, 0.6, 0.6},
		"tertiary_link":     {0.6, 0.6, 0.6},
		"unclassified_link": {0.6, 0.6, 0.6},
		"residential_link":  {0.6, 0.6, 0.6},
		"service_link":      {0.6, 0.6, 0.6},
	}
	x0, y0 := processLatLon(w.Nodes[0], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
	dc.MoveTo(x0, y0)
	for i := 1; i < len(w.Nodes); i++ {
		x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		dc.LineTo(x1, y1)
	}
	// set colors according to their type
	roadType := w.Way.Tags.Find("highway")
	if roadType == "" {
		return
	}
	c := colors[roadType]
	dc.SetColor(color.RGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), 255})
	dc.Stroke()
}

func drawBuilding(dc *gg.Context, w *WayWithNodes, centroidLat float64, centroidLon float64, factorX float64, factorY float64, canvasX float64, canvasY float64) {
	dc.SetLineWidth(1.0)
	for i := 0; i < len(w.Nodes)-1; i++ {
		x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		dc.LineTo(x1, y1)
	}
	dc.ClosePath()
	if w.Way.Tags.Find("building:levels") != "" {
		dc.SetLineWidth(5.0)
	}
	dc.SetRGB(1, 0, 0)
	dc.Stroke()
}
