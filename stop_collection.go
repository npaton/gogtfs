package gtfs

import (
	"math"
	"sort"
)

type StopDistanceResult struct {
	Stop   *Stop
	Distance float64
}

type StopDistanceResults []*StopDistanceResult

func (s StopDistanceResults) Len() int      { return len(s) }
func (s StopDistanceResults) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s StopDistanceResults) Less(i, j int) bool { return s[i].Distance < s[j].Distance }


type StopCollection struct {
	Stops  map[string]*Stop
	qt     *QuadTree
	maxLat float64
	maxLon float64
	minLat float64
	minLon float64
}

func NewStopCollection() StopCollection {
	return StopCollection{
		Stops:  make(map[string]*Stop),
		qt:     nil,
		maxLat: math.Inf(-1),
		maxLon: math.Inf(-1),
		minLat: math.Inf(1),
		minLon: math.Inf(1),
	}
}

func (c *StopCollection) Length() int {
	return len(c.Stops)
}

func (c *StopCollection) Stop(id string) (stop *Stop) {
	return c.Stops[id]
}

func (c *StopCollection) SetStop(id string, stop *Stop) {
	c.Stops[id] = stop
	c.maxLat = math.Max(c.maxLat, stop.Lat)
	c.maxLon = math.Max(c.maxLon, stop.Lon)
	c.minLat = math.Min(c.minLat, stop.Lat)
	c.minLon = math.Min(c.minLon, stop.Lon)
}

func (c *StopCollection) createQuadtree() {
	c.qt = CreateQuadtree(c.minLat, c.maxLat, c.minLon, c.maxLon)
	for _, stop := range c.Stops {
		c.qt.Insert(stop)
	}
}

func (c *StopCollection) StopsByName(name string) (results []*Stop) {
	for _, stop := range c.Stops {
		if stop.Name == name {
			results = append(results, stop)
		}
	}
	return
}

func (c *StopCollection) StopsByProximity(lat, lng, radius float64) (results []*Stop) {
	if c.qt == nil {
		c.createQuadtree()
	}
	return c.qt.SearchByProximity(lat, lng, radius)
}

func (c *StopCollection) StopDistancesByProximity(lat, lng, radius float64) (results StopDistanceResults) {
	stops := c.StopsByProximity(lat, lng, radius)
	stopdistances := make(StopDistanceResults, len(stops))
	for i, stop := range stops {
		stopdistances[i] = &StopDistanceResult{Stop: stop, Distance: stop.DistanceToCoordinate(lat, lng)}
	}
	sort.Sort(stopdistances)
	return stopdistances
}

func (c *StopCollection) RandomStop() (stopX *Stop) {
	stopsCount := 0
	for _, stopX = range c.Stops {
		if stopsCount > 10 {
			break
		}
		stopsCount += 1
	}
	return
}
