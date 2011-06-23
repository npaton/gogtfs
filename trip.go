package gtfs

import (
	"strconv"
	"time"
	"fmt"
	// "log"
)


// Trip.Direction possible values: (There is no "in" or "out" directions, they symbolize two opposite directions)
const (
	DirectionOut = iota // 0 - travel in one direction (e.g. outbound travel)
	DirectionIn         //  1 - travel in the opposite direction (e.g. inbound travel)
)

// trips.txt
type Trip struct {

	// route_id - Required. The route_id field contains an ID that uniquely identifies a route. 
	// This value is referenced from the routes.txt file.
	Route *Route

	// service_id - Required. The service_id contains an ID that uniquely identifies a set of dates when service 
	// is available for one or more routes. This value is referenced from the calendar.txt or calendar_dates.txt file.
	serviceId string

	// trip_id	- Required. The trip_id field contains an ID that identifies a trip. The trip_id is dataset unique.
	Id string

	// trip_headsign - Optional. The trip_headsign field contains the text that appears on a sign that identifies the 
	// trip's destination to passengers. Use this field to distinguish between different patterns of service in the 
	// same route. If the headsign changes during a trip, you can override the trip_headsign by specifying values 
	// for the the stop_headsign field in stop_times.txt.
	// See a Google Maps screenshot highlighting the headsign:
	// http://code.google.com/transit/spec/transit_feed_specification.html#transitTripHeadsignScreenshot
	Headsign string

	// trip_short_name	- Optional. The trip_short_name field contains the text that appears in schedules and sign boards 
	// to identify the trip to passengers, for example, to identify train numbers for commuter rail trips. If riders do 
	// not commonly rely on trip names, please leave this field blank.
	// A trip_short_name value, if provided, should uniquely identify a trip within a service day; it should not be used 
	// for destination names or limited/express designations.
	ShortName string

	// direction_id - Optional. The direction_id field contains a binary value that indicates the direction of travel for 
	// a trip. Use this field to distinguish between bi-directional trips with the same route_id. This field is not used 
	// in routing; it provides a way to separate trips by direction when publishing time tables. You can specify names for 
	// each direction with the trip_headsign field.
	// 		0 - travel in one direction (e.g. outbound travel)
	// 		1 - travel in the opposite direction (e.g. inbound travel)
	// For example, you could use the trip_headsign and direction_id fields together to assign a name to travel in each 
	// direction on trip "1234", the trips.txt file would contain these rows for use in time tables:
	// 
	// 		trip_id, ... ,trip_headsign,direction_id
	// 		1234, ... , to Airport,0
	// 		1505, ... , to Downtown,1
	// 
	// See DirectionIn/Out constants
	Direction byte

	// block_id - Optional. The block_id field identifies the block to which the trip belongs. A block consists of two
	// or more sequential trips made using the same vehicle, where a passenger can transfer from one trip to the next just
	// by staying in the vehicle. The block_id must be referenced by two or more trips in trips.txt.
	BlockId string

	// shape_id - Optional. The shape_id field contains an ID that defines a shape for the trip. This value is referenced
	// from the shapes.txt file. The shapes.txt file allows you to define how a line should be drawn on the map to represent a trip.
	ShapeId string

	//
	DayRange

	StopTimes []*StopTime
	
	Frequencies []Frequency
	

	feed *Feed
}

func (t *Trip) NextStopTimeWithTransfer(fromstop, after *Stop)  (stopTime *StopTime, cost int, doesRun bool) {
	// Looking for stop after fromstop in trip sequence that has attached stoptimes
	if fromstop == nil { return nil, 0, false }
	foundFrom  := false
	foundAfter := false
	var previousStopTime *StopTime
	for _, st := range t.StopTimes {
		if foundFrom {
			// We from departure to departure (and not arrival) to automatically include transfer/wait time. Kinda.
			cost += int(st.DepartureTime - previousStopTime.DepartureTime)
			previousStopTime = st
		}
		if st.Stop != nil {
			if !foundFrom && st.Stop == fromstop {
				foundFrom = true
				previousStopTime = st
			}
			if len(st.Stop.StopTimes) > 1 && foundFrom && (after == nil || foundAfter) {
				return st, cost, true
			}
			if after != nil && !foundAfter && st.Stop == after {
				foundAfter = true
			}
		}
	}
	return nil, 0, false
}

func (t *Trip) RunsFromTo(fromstop, tostop *Stop) (stopTime *StopTime, cost int, doesRun bool) {
	if fromstop == nil { return nil, 0, false }
	if tostop == nil { return nil, 0, false }
	
	foundFrom := false
	var previousStopTime *StopTime
	for _, st := range t.StopTimes {
		if foundFrom {
			// We from departure to departure (and not arrival) to automatically include transfer/wait time. Kinda.
			cost += int(st.DepartureTime - previousStopTime.DepartureTime)
			previousStopTime = st
		}
		if !foundFrom && st.Stop != nil && st.Stop == fromstop {
			foundFrom = true
			previousStopTime = st
		}
		if st.Stop != nil && st.Stop.Id == tostop.Id {
			if !foundFrom {
				return nil, 0, false
			}
			return st, cost, true
		}
	}
	return nil, 0, false
}
func (t *Trip) RunsAccross(stop *Stop) bool {
	
	for _, st := range t.StopTimes {
		if st.Stop.Id == stop.Id {
			return true
		}
	}
	return false
}

// AddStopTime adds StopTime to trip.StopTimes with respect to the stop_sequence order
func (t *Trip) AddStopTime(newStopTime *StopTime) {
	if t.StopTimes == nil {
		t.StopTimes = make([]*StopTime, 0, 5)
	}

	stopTimesLength := len(t.StopTimes)
	if stopTimesLength == 0 {
		// If first element simply append
		t.StopTimes = append(t.StopTimes, newStopTime)
	} else {
		// If new stop time stop seq is superior to the last one on StopTimes array, append
		if t.StopTimes[len(t.StopTimes)-1].StopSequence < newStopTime.StopSequence {
			t.StopTimes = append(t.StopTimes, newStopTime)
		} else {
			// Otherwise rebuild new array, inserting the new stop time at right time
			newStopTimes := make([]*StopTime, stopTimesLength+1)
			hasAppendedNewStopTime := false
			for _, existingStopTime := range t.StopTimes {
				if existingStopTime != nil {
					if !hasAppendedNewStopTime && newStopTime.StopSequence < existingStopTime.StopSequence {
						newStopTimes = append(newStopTimes, newStopTime)
						hasAppendedNewStopTime = true
					}
					newStopTimes = append(newStopTimes, existingStopTime)
				}
			}
			t.StopTimes = newStopTimes
		}
	}
}

type DayRange struct {
	from uint // time of day in seconds since midnight
	to uint // in seconds
}

func (a *DayRange) Intersects(b *DayRange) bool {
	return (a.from <= b.from && a.to >= b.from) || (b.from <= a.from && b.to >= a.from)
}

func (a *DayRange) Contains(b *DayRange) bool {
	return a.from <= b.from && a.to >= b.to
}

func (a *DayRange) Add(b *DayRange) {
	if b.from < a.from {
		a.from = b.from
	}
	if b.to > a.to {
		a.to = b.to
	}
}

func (t *Trip) calculateDayTimeRange() {
	stopTimesLength := len(t.StopTimes)
	if stopTimesLength > 0 {
		dayrange := &DayRange{t.StopTimes[0].DepartureTime, t.StopTimes[stopTimesLength-1].ArrivalTime}
		for _, freq := range t.Frequencies {
			dayrange.Add(&freq.DayRange)
		}
		t.DayRange = *dayrange
	} else {
		t.DayRange = DayRange{0, 0}
	}

}

func (t *Trip) HasShape() bool {
	return t.ShapeId != "" && t.feed.Shapes[t.ShapeId] != nil
}

func (t *Trip) RunsOn(date *time.Time) (runs bool) {
	runs = false // Unnecessary, default init to false, no?
	intdate, _ := strconv.Atoi(fmt.Sprintf("%04d", date.Year) + fmt.Sprintf("%02d",date.Month) + fmt.Sprintf("%02d",date.Day)) 
	if calendar, ok := t.feed.Calendars[t.serviceId]; ok {
		if calendar.ValidOn(intdate, date) {
			runs = true
		}
	}

	if calendardates, ok := t.feed.CalendarDates[t.serviceId]; ok {
		// log.Println("calendardates", calendardates)
		for _, cd := range calendardates {
			if exceptionOnDay, shouldRun := cd.ExceptionOn(intdate); exceptionOnDay {
				// log.Println("calendardate shouldRun", shouldRun)
				runs = shouldRun
			}
		}
	}

	return
}

func (t *Trip) afterInit() {
	t.Frequencies = make([]Frequency, 0)
}


func (t *Trip) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "trip_id":
		t.Id = val
		break
	case "route_id":
		t.Route = t.feed.Routes[val]
		break
	case "service_id":
		t.serviceId = val
		break
	case "trip_headsign":
		t.Headsign = val
		break
	case "trip_short_name":
		t.ShortName = val
		break
	case "direction_id":
		v, _ := strconv.Atoui(val) // Should panic on error !
		if v == 0 {
			t.Direction = DirectionOut
		} else if v == 1 {
			t.Direction = DirectionIn
		}
		break
	case "block_id":
		t.BlockId = val
		break
	case "shape_id":
		t.ShapeId = val
		break
	}
}

func (t *Trip) copyColorToShape() {
	if t.Route == nil {
		return
	}
	color := t.Route.Color
	if color != "" && t.feed.Shapes[t.ShapeId] != nil {
		// log.Println("INSPECT", color, "===", t.feed.Shapes[t.ShapeId])
		t.feed.Shapes[t.ShapeId].Color = color
	} else if t.feed.Shapes[t.ShapeId] != nil {
		t.feed.Shapes[t.ShapeId].Color = "000000"
	}
	// log.Println("INSPECT", color, "===", t.feed.Shapes[t.ShapeId])
}
