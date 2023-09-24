# OSM-viz
This is just a simple program that grabs a .osm file and prints it in a "beautiful" png

It is not very optimized and the code is very poorly documented. I hope I can clean it up in the future.

For a 76K node map it takes around 8s in an i7 

Planned improvements:
- Palette
- cmd line arguments
- Make it run at least 10x faster
- Make the parser run 1 time per osm, that is comb thru the nodes only one time
- Add cars, trees and other stuff
- Isochrone


To run:

```bash
go run -v ./main.go
```