package gtfs

import (
	// "time"
	"strconv"
)

// frequencies.txt
// The frequencies file is optional. This table is intended to represent schedules that don't have a fixed list of stop times. 
// When trips are defined in frequencies.txt, the trip planner ignores the absolute values of the arrival_time and departure_time 
// fields for those trips in stop_times.txt. Instead, the stop_times table defines the sequence of stops and the time difference 
// between each stop.
type Frequency struct {

	// trip_id	- Required. The trip_id contains an ID that identifies a trip on which the specified frequency of service applies. 
	// Trip IDs are referenced from the trips.txt file.
	Trip *Trip

	// start_time - Required. The start_time field specifies the time at which service begins with the specified frequency. 
	// The time is measured from "noon minus 12h" (effectively midnight, except for days on which daylight savings time changes occur) 
	// at the beginning of the service date. For times occurring after midnight, enter the time as a value greater than 24:00:00 in 
	// HH:MM:SS local time for the day on which the trip schedule begins. E.g. 25:35:00.
	StartTime uint

	// end_time - Required. The end_time field indicates the time at which service changes to a different frequency (or ceases) 
	// at the first stop in the trip. The time is measured from "noon minus 12h" (effectively midnight, except for days on which 
	// daylight savings time changes occur) at the beginning of the service date. For times occurring after midnight, enter the 
	// time as a value greater than 24:00:00 in HH:MM:SS local time for the day on which the trip schedule begins. E.g. 25:35:00.
	EndTime uint

	// headway_secs - Required. The headway_secs field indicates the time between departures from the same stop (headway) for this 
	// trip type, during the time interval specified by start_time and end_time. The headway value must be entered in seconds.
	// 
	// Periods in which headways are defined (the rows in frequencies.txt) shouldn't overlap for the same trip, since it's hard to 
	// determine what should be inferred from two overlapping headways. However, a headway period may begin at the exact same time 
	// that another one ends, for instance:
	// 		A, 05:00:00, 07:00:00, 600
	// 		B, 07:00:00, 12:00:00, 1200
	HeadwaySecs uint

	DayRange
	feed *Feed
}

func (f *Frequency) calculateDayTimeRange() {
	f.DayRange = DayRange{f.StartTime, f.EndTime}
}

func (f *Frequency) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "trip_id":
		f.Trip = f.feed.Trips[val]
		break
	case "start_time":
		v, err := timeOfDayStringToSeconds(val)
		if err != nil {
			panic(err.Error() + val)
		}
		f.StartTime = v
		break
	case "end_time":
		v, err := timeOfDayStringToSeconds(val)
		if err != nil {
			panic(err)
		}
		f.EndTime = v
		break
	case "headway_secs":
		v, _ := strconv.Atoi(val)
		f.HeadwaySecs = uint(v)
		break
	}
}
