package gtfs

type StopCollection struct {
	Stops map[string]*Stop
}

func NewStopCollection() StopCollection {
	return StopCollection{make(map[string]*Stop)}
}

func (c *StopCollection) Length() int {
	return len(c.Stops)
}

func (c *StopCollection) Stop(id string) (stop *Stop) {
	return c.Stops[id]
}

func (c *StopCollection) SetStop(id string, stop *Stop) {
	c.Stops[id] = stop
}

func (c *StopCollection) StopsByName(name string) (results []*Stop) {
	for _, stop := range c.Stops {
		if stop.Name == name {
			results = append(results, stop)
		}
	}
	return
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
