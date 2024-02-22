package main

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/paulmach/osm"
)

type WayWithNodes struct {
	Way   *osm.Way
	Nodes []*osm.Node
}

type mapData struct {
	ways        []*WayWithNodes
	relations   []*osm.Relation
	centroidLat float64
	centroidLon float64
	factorX     float64
	factorY     float64
	canvasX     float64
	canvasY     float64
}

func main() {
	start := time.Now()

	nodeMap, osmMap := loadNodeMapFromFile("lima.osm")
	fmt.Printf("Mapped all nodes %v, nodes: %d \n", time.Since(start), len(nodeMap))

	mapData := getMapData(nodeMap, osmMap)
	// clean canvas
	dc := gg.NewContext(int(mapData.canvasX), int(mapData.canvasY))
	// dc.SetRGB(254/255.0, 255.0/255.0, 229.0/255.0)
	dc.SetRGB(147/255.0, 242.0/255.0, 255.0/255.0) // blue
	// dc.SetRGB(0.9, 0.9, 0.8)
	dc.DrawRectangle(0, 0, mapData.canvasX, mapData.canvasY)
	dc.Fill()
	fmt.Println("Done preprocessing in " + time.Since(start).String())

	drawDistricts(dc, mapData)

	// drawRoadsAndBuildings(dc, mapData)
	// fmt.Println("Drawing buildings in " + time.Since(start).String())

	timestamp := time.Now().UnixMicro()

	fmt.Println("Saving PNG in " + time.Since(start).String())
	dc.SavePNG(fmt.Sprintf("map-%d.png", timestamp))

	fmt.Println("Done in " + time.Since(start).String())
}

type Way struct {
	id       string
	start    *osm.Node
	end      *osm.Node
	nodes    []*osm.Node
	reversed bool
}

func drawDistricts(dc *gg.Context, mapData *mapData) {
	for _, r := range mapData.relations {
		fmt.Println(r.Tags.Find("name"))
		// if r.Tags.Find("name") != "La Molina" {
		// 	continue
		// }
		ways := []Way{}
		for _, m := range r.Members {
			if m.Type == "way" {
				for _, w := range mapData.ways {
					if w.Way.ID == m.ElementID().WayID() {
						way := Way{
							id:    m.ElementID().String(),
							start: w.Nodes[0],
							end:   w.Nodes[len(w.Nodes)-1],
							nodes: w.Nodes,
						}
						ways = append(ways, way)
					}
				}
			}
		}
		sortedWays := []Way{}
		visitedWays := map[string]bool{}
		start := ways[0]
		visitedWays[start.id] = true

		distance := func(a, b *osm.Node) float64 {
			return (a.Lat-b.Lat)*(a.Lat-b.Lat) + (a.Lon-b.Lon)*(a.Lon-b.Lon)
		}

		// sort nodes by distance
		var last *osm.Node

		sortedWays = append(sortedWays, start)
		last = start.end
		visitedWays[start.id] = true

		for i := 0; i < len(ways); i++ {
			minDist := 1000000.0
			var cand *Way
			for _, w := range ways {
				w := w
				if visitedWays[w.id] {
					continue
				}
				if distance(last, w.start) < minDist {
					minDist = distance(last, w.start)
					cand = &w
				}
				if distance(last, w.end) < minDist {
					minDist = distance(last, w.end)
					w.reversed = true
					cand = &w
				}
			}
			if cand != nil {
				sortedWays = append(sortedWays, *cand)
				visitedWays[cand.id] = true
				if cand.reversed {
					last = cand.start
				} else {
					last = cand.end
				}
			}
		}

		// draw the districts
		dc.NewSubPath()
		dc.Push()
		startX, startY := 0.0, 0.0
		for i := 0; i < len(sortedWays); i++ {
			nodes := sortedWays[i].nodes
			if sortedWays[i].reversed {
				aux := []*osm.Node{}
				for j := len(nodes) - 1; j >= 0; j-- {
					aux = append(aux, nodes[j])
				}
				nodes = aux
			}
			for _, n := range nodes {
				x1, y1 := processLatLon(n, mapData)
				if i == 0 && startX == 0 && startY == 0 {
					startX, startY = x1, y1
					dc.DrawCircle(x1, y1, 5)
					dc.MoveTo(x1, y1)
				} else {
					dc.LineTo(x1, y1)
				}
			}
		}

		c := [][3]float64{
			{0, 128, 128},
			{112, 164, 148},
			{180, 200, 168},
			{246, 237, 189},
			{237, 187, 138},
			{222, 137, 90},
			{200, 86, 44},
			{112, 164, 148},
			{180, 200, 168},
			{246, 237, 189},
			{237, 187, 138},
			{222, 138, 90},
			{202, 86, 44},
		}
		random := rand.Intn(len(c))
		// random := i % len(c)
		color := color.NRGBA{
			uint8(c[random][0]),
			uint8(c[random][1]),
			uint8(c[random][2]),
			255}
		dc.SetColor(color)
		dc.SetFillStyle(gg.NewSolidPattern(color))
		dc.ClosePath()
		dc.Fill()
		dc.Pop()

		dc.NewSubPath()
		dc.Push()

		startX, startY = 0.0, 0.0
		for i := 0; i < len(sortedWays); i++ {
			nodes := sortedWays[i].nodes
			if sortedWays[i].reversed {
				aux := []*osm.Node{}
				for j := len(nodes) - 1; j >= 0; j-- {
					aux = append(aux, nodes[j])
				}
				nodes = aux
			}
			for _, n := range nodes {
				x1, y1 := processLatLon(n, mapData)
				if i == 0 && startX == 0 && startY == 0 {
					startX, startY = x1, y1
					dc.DrawCircle(x1, y1, 5)
					dc.MoveTo(x1, y1)
				} else {
					dc.LineTo(x1, y1)
				}
			}
		}

		dc.LineTo(startX, startY)
		dc.SetLineWidth(20)
		dc.SetRGB(254/255.0, 255.0/255.0, 229.0/255.0)
		// dc.SetRGB(0/255.0, 0.0/255.0, 0.0/255.0)
		dc.Stroke()
		dc.Pop()
	}
}

func loadNodeMapFromFile(filename string) (map[osm.NodeID]*osm.Node, *osm.OSM) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the file using paulmach/osm
	bytes, _ := io.ReadAll(file)
	mappy := &osm.OSM{}
	xml.Unmarshal(bytes, mappy)

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

	return nodeMap, mappy
}

func getMapData(nodeMap map[osm.NodeID]*osm.Node, osmMap *osm.OSM) *mapData {
	var (
		minLat = 1000.0
		maxLat = -1000.0
		minLon = 1000.0
		maxLon = -1000.0
	)

	canvasX := 5 * 1000.0
	canvasY := 5 * 1000.0

	// Get ways, with centroids
	centroidLat := 0.0
	centroidLon := 0.0
	count := 0

	relations := []*osm.Relation{}
	for _, r := range osmMap.Relations {
		if r.Tags.Find("type") == "boundary" &&
			r.Tags.Find("boundary") == "administrative" &&
			r.Tags.Find("admin_level") == "8" {
			relations = append(relations, r)
		}
	}
	ways := []*WayWithNodes{}
	for _, w := range osmMap.Ways {
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
	zoom := 3.0 // zoom 3:0
	factorX := canvasX / (maxLat - minLat) * zoom
	factorY := canvasY / (maxLon - minLon) * zoom

	println(factorX, factorY)

	return &mapData{
		centroidLat: centroidLat - 0.05, // up down
		centroidLon: centroidLon - 0.05, // left right
		factorX:     factorX,
		factorY:     factorY,
		canvasX:     canvasX,
		canvasY:     canvasY,
		ways:        ways,
		relations:   relations,
	}
}

func ToNode(n osm.WayNode) *osm.Node {
	return &osm.Node{
		ID:  n.ID,
		Lat: n.Lat,
		Lon: n.Lon,
	}
}

func drawNodes(dc *gg.Context, nodes []*osm.Node, mapData *mapData) {
	x0, y0 := processLatLon(nodes[0], mapData)
	dc.MoveTo(x0, y0)
	for i := 0; i < len(nodes); i++ {
		x1, y1 := processLatLon(nodes[i], mapData)
		dc.LineTo(x1-2, y1-2)
	}
}

func drawRoadsAndBuildings(dc *gg.Context, mapData *mapData) {

	var buildings []*WayWithNodes = []*WayWithNodes{}

	// sort buildings by latitude (this bubble sort, not very efficient)
	for i := 0; i < len(buildings); i++ {
		for j := i + 1; j < len(buildings); j++ {
			if buildings[i].Nodes[0].Lat < buildings[j].Nodes[0].Lat {
				buildings[i], buildings[j] = buildings[j], buildings[i]
			}
		}
	}
	for _, w := range buildings {
		drawBuilding(dc, w, mapData)
	}

	dc.Fill()
	for _, w := range mapData.ways {
		// continue
		if len(w.Nodes) == 0 {
			continue
		}
		dc.NewSubPath()
		if w.Way.Tags.AnyInteresting() {
			if w.Way.Tags.Find("highway") != "" {
				drawRoad(dc, w, mapData)
				// for _, t := range w.Way.Tags {
				// 	println(t.Key, t.Value)
				// }
			}
			if w.Way.Tags.Find("leisure") != "" {

				dc.SetLineWidth(1.0)
				dc.Push()
				dc.NewSubPath()
				for i := 0; i < len(w.Nodes)-1; i++ {
					x1, y1 := processLatLon(w.Nodes[i], mapData)
					if i == 0 {
						dc.MoveTo(x1, y1)
					} else {
						dc.LineTo(x1, y1)
					}
				}
				dc.ClosePath()
				// 106, 153, 78
				// 180, 219, 213
				dc.SetRGB(205/255.0, 239.0/255.0, 212.0/255.0)
				dc.Fill()
				dc.Pop()
			}
			if w.Way.Tags.Find("building") != "" {
				buildings = append(buildings, w)
			}
		} else {
			continue
		}
	}
}
