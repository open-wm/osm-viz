package main

import "github.com/paulmach/osm"

func processLatLon(node *osm.Node, mapData *mapData) (float64, float64) {
	x0 := (node.Lon-mapData.centroidLon)*mapData.factorX + mapData.canvasX/2
	lambda := 100.0
	skew := (1.0 - x0/(mapData.canvasY/2.0)) * lambda
	skew = 0.0
	shrinkage := 0.8
	y0 := (node.Lat-mapData.centroidLat)*-1.0*mapData.factorY*shrinkage + mapData.canvasY/2 + skew

	// angle := math.Pi / 2.0
	// // rotate x and y around the centroid
	// tempX0 := x0
	// tempY0 := y0
	// x0 = (tempX0-canvasX/2)*math.Cos(angle) - (tempY0-canvasY/2)*math.Sin(angle) + canvasX/2
	// y0 = (tempX0-canvasX/2)*math.Sin(angle) + (tempY0-canvasY/2)*math.Cos(angle) + canvasY/2
	// y0 *= 0.5

	return x0, y0
}
