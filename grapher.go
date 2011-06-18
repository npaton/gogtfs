package gtfs

type RouteConnections struct {
	connections map[string][]*RouteConnection
} 

func NewRouteConnections() *RouteConnections {
	return &RouteConnections{make(map[string][]*RouteConnection, 100000)}
}

type RouteConnection struct {
	*Route
	*Stop
}

func (a *RouteConnection)Equal(b *RouteConnection) bool {
	return a.Route.Id == b.Route.Id && a.Stop.Id == b.Stop.Id
}


func (rc *RouteConnections)collectRouteConnections(trips []*Trip) (count int) {
	for _, trip := range trips {
		route := trip.Route
		rc.connections[route.Id] = make([]*RouteConnection, 0, len(trip.StopTimes))
		for _, newcon := range trip.ConnectedRoutes {
			shouldAdd := true
			for _, otherconn := range rc.connections[route.Id] {
				if newcon.Equal(otherconn) {
					shouldAdd = false
					break
				}
			}
			if shouldAdd {
				count += 1
				rc.connections[route.Id] = append(rc.connections[route.Id], newcon)
			}
		}
	}
	return
}