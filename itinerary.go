package gtfs

// import (
// 	"time"
// )
// 
// type Itinerary struct {
// 	From *Stop
// 	To *Stop
// 	
// 	MaxTransfers int // -1 = unlimited. Default: 3
// 	MaxDuration int // In seconds. -1 = unlimited. Default: 60*60*3 (3 hours)
// 	
// 	feed *Feed
// 	trips []*Trip
// 	bestTimeAtStop map[*Stop]int
// 	
// }
// 
// type Segment struct {
// 	From *Stop
// 	To *Stop
// 	DepartureTime *time.Time
// 	ArrivalTime *time.Time
// 	Steps []Step
// }
// 
// type Step struct {
// 	to *StopTime // stop time index in trip
// 	previous *Step
// }
// 
// type Stepper struct {
// 	previous *StopTime
// 	itinerary *Itinerary
// }
// 
// func NewItinerary(f *Feed) (i *Itinerary) {
// 	i = &Itinerary{}
// 	i.MaxDuration = 60*60*3
// 	i.MaxTransfers = 3
// 	i.Step = Step{Steps:make([]Step,0)}
// 	i.feed = f
// 	return 
// }
// 
// func (i *Itinerary)Run() {
// 	i.trips = i.feed.TripsForDayAndDayRange(i.DepartureTime, &DayRange{uint(i.DepartureTime.Hour), 60*30*3})
// 	branchCountChan := make(chan bool, 10)
// 	branchCompletedChan := make(chan *Step, 10)
// 	
// 	// start off the fun HERE
// 	
// 	branchesCount := 0
// 	select {
// 	case <- branchCountChan:
// 		// Count all search branching
// 		branchesCount = branchesCount + 1
// 	case step := <- branchCompletedChan:
// 		if step.to.Stop.Id == i.To.Id {
// 			log.Println("Found route!!", step.cost)
// 		}
// 		
// 		// Uncount branches and when all branches have ended, break
// 		branchesCount = branchesCount -1
// 		if branchesCount <= 0 {
// 			break
// 		}
// 	case time.Sleep(10*1E9): // Timout after 10s
// 		panic("Itinerary more than 10s to complete")
// 	}
// }
// 
// func (i *Itinerary)step(stepper Stepper) {
// 	go func() {
// 		// trips = tripsIntersectingStop(trips, stop)
// 		// for _, trip := range trips {
// 		// 	if trip.RunsAccross(i.To) {
// 		stepper.Step // Do shit
// 	}
// }
// 
// 
// 
// func tripsIntersectingStop(trips []*Trip, stop *Stop) []*Trip {
// 	foundTrips := make([]*Trip, 0, len(trips))
// 	for _, trip := range trips {
// 		if trip.RunsAccross(stop) {
// 			foundTrips = append(foundTrips, trip)
// 		}
// 	}
// 	return foundTrips
// }