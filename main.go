package main

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/paulmach/osm"
)

type WayWithNodes struct {
	Way   *osm.Way
	Nodes []*osm.Node
}

func main() {
	start := time.Now()
	file, err := os.Open("map-sani.osm")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the file using paulmach/osm
	bytes, _ := io.ReadAll(file)
	mappy := &osm.OSM{}
	xml.Unmarshal(bytes, mappy)

	// Print the number of nodes

	canvasX := 10000.0
	canvasY := 5000.0
	dc := gg.NewContext(int(canvasX), int(canvasY))

	centroidLat := 0.0
	centroidLon := 0.0
	count := 0
	minLat := 1000.0
	maxLat := -1000.0
	minLon := 1000.0
	maxLon := -1000.0

	nodeMap := map[osm.NodeID]*osm.Node{}

	var wg sync.WaitGroup

	for i := 0; i < 1; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			for _, n := range mappy.Nodes[i*len(mappy.Nodes)/1 : (i+1)*len(mappy.Nodes)/1] {
				nodeMap[n.ID] = n
			}
		}()
	}
	wg.Wait()

	fmt.Println("Mapped all nodes " + time.Since(start).String())

	ways := []*WayWithNodes{}

	for _, w := range mappy.Ways {
		nway := WayWithNodes{Way: w}
		for _, n := range w.Nodes {
			node, ok := nodeMap[n.ID]
			if !ok {
				continue
			}
			nway.Nodes = append(nway.Nodes, node)
			if node.Lat < minLat {
				minLat = node.Lat
			}
			if node.Lat > maxLat {
				maxLat = node.Lat
			}
			if n.Lon < minLon {
				minLon = node.Lon
			}
			if node.Lon > maxLon {
				maxLon = node.Lon
			}
			centroidLat += node.Lat
			centroidLon += node.Lon
			count++
		}
		ways = append(ways, &nway)
	}
	centroidLat /= float64(count)
	centroidLon /= float64(count)
	println(centroidLat, centroidLon, minLat, maxLat, minLon, maxLon)
	zoom := 2.0
	factorX := canvasX / (maxLat - minLat) * zoom
	factorY := canvasY / (maxLat - minLat) * zoom

	fmt.Println("Done preprocessing in " + time.Since(start).String())
	println(factorX, factorY)

	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, canvasX, canvasY)
	dc.Fill()

	for _, w := range ways {
		if len(w.Nodes) == 0 {
			continue
		}
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

	fmt.Println("Saving PNG in " + time.Since(start).String())
	dc.SavePNG(fmt.Sprintf("map-%d.png", timestamp))

	fmt.Println("Done in " + time.Since(start).String())
}

func processLatLon(node *osm.Node, centroidLat float64, centroidLon float64, factorX float64, factorY float64, canvasX float64, canvasY float64) (float64, float64) {
	x0 := (node.Lon-centroidLon)*factorX + canvasX/2
	lambda := 100.0
	skew := (1.0 - x0/(canvasY/2.0)) * lambda
	skew = 0.0
	shrinkage := 0.8
	y0 := (node.Lat-centroidLat)*-1.0*factorY*shrinkage + canvasY/2 + skew
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
		lineWidth = laneFloat*5.0 + 1.0
	}
	dc.SetLineWidth(lineWidth)
	colors := map[string][3]float64{
		"motorway":          {251 / 255.0, 133.0 / 255.0, 0.0 / 255.0}, // Javier prado y evitamiento // orange
		"trunk":             {2 / 255.0, 48.0 / 255.0, 71.0 / 255.0},   // IDK we dont seem to have any
		"primary":           {251 / 255.0, 183.0 / 255.0, 0.0 / 255.0}, // Panamericana // yellow
		"secondary":         {2 / 255.0, 48.0 / 255.0, 71.0 / 255.0},   // la floresta, olguin // blue
		"tertiary":          {33 / 255.0, 157 / 255.0, 188 / 255.0},
		"unclassified":      {0.6, 0.6, 0.6},
		"residential":       {0.6, 0.6, 0.6},
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
	x0, y0 := processLatLon(w.Nodes[0], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
	dc.MoveTo(x0, y0)
	for i := 1; i < len(w.Nodes); i++ {
		x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		// if y1 < canvasY/3 || y1 > canvasY*2/3 {
		// 	continue
		// }
		dc.LineTo(x1, y1)
	}

	// set colors according to their type
	roadType := w.Way.Tags.Find("highway")
	if roadType == "" {
		return
	}
	c := colors[roadType]
	if c[0] == 0.0 && c[1] == 0.0 && c[2] == 0.0 {
		dc.SetRGBA(0.6, 0.6, 0.6, 0.3)
	} else {
		dc.SetColor(color.NRGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), 200})
	}
	dc.Stroke()
}

func drawBuilding(dc *gg.Context, w *WayWithNodes, centroidLat float64, centroidLon float64, factorX float64, factorY float64, canvasX float64, canvasY float64) {
	dc.SetLineWidth(1.0)
	for i := 0; i < len(w.Nodes)-1; i++ {
		x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		dc.LineTo(x1, y1)
	}
	dc.ClosePath()
	dc.SetRGB(1, 0, 0)
	dc.Stroke()

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
			height = float64(level) * 20
		}
		dc.SetRGBA(0.5, 0.5, 0.8, 1)
		dc.SetLineWidth(1.0)
	} else {
		return
	}
	for i := 0; i < len(w.Nodes)-1; i++ {
		x1, y1 := processLatLon(w.Nodes[i], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		dc.MoveTo(x1, y1)
		// for each node in the building lets draw a line that goes up
		dc.LineTo(x1, y1-height)
		x2, y2 := processLatLon(w.Nodes[i+1], centroidLat, centroidLon, factorX, factorY, canvasX, canvasY)
		dc.LineTo(x2, y2-height)
		dc.Stroke()

		// add a fill
		// color := dc.
		dc.Push()
		dc.SetRGB(1, 0, 0)

		start := min(y1, y2)
		end := max(y1, y2)
		// startX := min(x1, x2)
		// endX := max(x1, x2)

		linear := gg.NewLinearGradient(0, start-height, 0, end+height)
		linear.AddColorStop(0, color.RGBA{255, 0, 0, 255})
		linear.AddColorStop(1, color.RGBA{255, 255, 255, 255})
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
		// dc.
	}
}
