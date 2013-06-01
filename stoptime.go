package gtfs

import (
	"strconv"
	"strings"
	"time"
)

// StopTime.PickupType possible values:
const (
	PickupRegular     = iota // 0 - Regularly scheduled pickup
	PickupUnavailable        // 1 - No pickup available
	PickupThePhone           // 2 - Must phone agency to arrange pickup
	PickupTheDriver          // 3 - Must coordinate with driver to arrange pickup
)

// StopTime.DropOffType possible values:
const (
	DropOffRegular     = iota // 0 - Regularly scheduled drop off
	DropOffUnavailable        // 1 - No drop off available
	DropOffThePhone           // 2 - Must phone agency to arrange drop off
	DropOffTheDriver          // 3 - Must coordinate with driver to arrange drop off
)

// stop_times.txt
type StopTime struct {

	// Required. The trip_id field contains an ID that identifies a trip. This value is referenced from the trips.txt file.
	Trip *Trip

	// arrival_time - Required. The arrival_time specifies the arrival time at a specific stop for a specific trip on a route. 
	// The time is measured from "noon minus 12h" (effectively midnight, except for days on which daylight savings time changes occur) 
	// at the beginning of the service date. For times occurring after midnight on the service date, enter the time as a value greater 
	// than 24:00:00 in HH:MM:SS local time for the day on which the trip schedule begins. If you don't have separate times for arrival 
	// and departure at a stop, enter the same value for arrival_time and departure_time.
	// 
	// You must specify arrival times for the first and last stops in a trip. If this stop isn't a time point, use an empty string value 
	// for the arrival_time and departure_time fields. Stops without arrival times will be scheduled based on the nearest preceding timed 
	// stop. To ensure accurate routing, please provide arrival and departure times for all stops that are time points. 
	// Do not interpolate stops.
	// 
	// Times must be eight digits in HH:MM:SS format (H:MM:SS is also accepted, if the hour begins with 0). Do not pad times with spaces. 
	// The following columns list stop times for a trip and the proper way to express those times in the arrival_time field:
	// 		Time         	arrival_time value
	// 		08:10:00 A.M.	08:10:00 or 8:10:00
	// 		01:05:00 P.M.	13:05:00
	// 		07:40:00 P.M.	19:40:00
	// 		01:55:00 A.M.	25:55:00
	// 
	// Note: Trips that span multiple dates will have stop times greater than 24:00:00. For example, if a trip begins at 10:30:00 p.m. 
	// and ends at 2:15:00 a.m. on the following day, the stop times would be 22:30:00 and 26:15:00. Entering those stop times as 22:30:00 
	// and 02:15:00 would not produce the desired results.
	ArrivalTime uint

	// departure_time - Required. The departure_time specifies the departure time from a specific stop for a specific trip on a route. 
	// The time is measured from "noon minus 12h" (effectively midnight, except for days on which daylight savings time changes occur) 
	// at the beginning of the service date. For times occurring after midnight on the service date, enter the time as a value greater 
	// than 24:00:00 in HH:MM:SS local time for the day on which the trip schedule begins. If you don't have separate times for arrival 
	// and departure at a stop, enter the same value for arrival_time and departure_time.
	// 
	// You must specify departure times for the first and last stops in a trip. If this stop isn't a time point, use an empty string value
	// for the arrival_time and departure_time fields. Stops without arrival times will be scheduled based on the nearest preceding timed
	// stop. To ensure accurate routing, please provide arrival and departure times for all stops that are time points. 
	// Do not interpolate stops.
	// Times must be eight digits in HH:MM:SS format (H:MM:SS is also accepted, if the hour begins with 0). Do not pad times with spaces. 
	// The following columns list stop times for a trip and the proper way to express those times in the departure_time field:
	// => See ArrivalTime Table and Note.
	DepartureTime uint

	// stop_id - Required. The stop_id field contains an ID that uniquely identifies a stop. Multiple routes may use the same stop. 
	// The stop_id is referenced from the stops.txt file. If location_type is used in stops.txt, all stops referenced in stop_times.txt 
	// must have location_type of 0.
	// Where possible, stop_id values should remain consistent between feed updates. In other words, stop A with stop_id 1 should have 
	// stop_id 1 in all subsequent data updates. If a stop is not a time point, enter blank values for arrival_time and departure_time.
	Stop *Stop

	// stop_sequence - Required. The stop_sequence field identifies the order of the stops for a particular trip. The values for 
	// stop_sequence must be non-negative integers, and they must increase along the trip.
	// For example, the first stop on the trip could have a stop_sequence of 1, the second stop on the trip could have a stop_sequence 
	// of 23, the third stop could have a stop_sequence of 40, and so on.
	StopSequence uint

	// stop_headsign - Optional. The stop_headsign field contains the text that appears on a sign that identifies the trip's destination 
	// to passengers. Use this field to override the default trip_headsign when the headsign changes between stops. If this headsign is 
	// associated with an entire trip, use trip_headsign instead.
	// See a Google Maps screenshot highlighting the headsign: 
	// http://code.google.com/transit/spec/transit_feed_specification.html#transitTripHeadsignScreenshot
	Headsign string

	// pickup_type - Optional. The pickup_type field indicates whether passengers are picked up at a stop as part of the normal schedule
	// or whether a pickup at the stop is not available. This field also allows the transit agency to indicate that passengers must call 
	// the agency or notify the driver to arrange a pickup at a particular stop. Valid values for this field are:
	// 		0 - Regularly scheduled pickup
	// 		1 - No pickup available
	// 		2 - Must phone agency to arrange pickup
	// 		3 - Must coordinate with driver to arrange pickup
	// The default value for this field is 0.
	// See Pickup constants
	PickupType byte

	// drop_off_type - Optional. The drop_off_type field indicates whether passengers are dropped off at a stop as part of the normal 
	// schedule or whether a drop off at the stop is not available. This field also allows the transit agency to indicate that passengers 
	// must call the agency or notify the driver to arrange a drop off at a particular stop. Valid values for this field are:
	// 		0 - Regularly scheduled drop off
	// 		1 - No drop off available
	// 		2 - Must phone agency to arrange drop off
	// 		3 - Must coordinate with driver to arrange drop off
	// The default value for this field is 0.
	// See DropOff constants
	DropOffType byte

	// shape_dist_traveled - Optional. When used in the stop_times.txt file, the shape_dist_traveled field positions a stop as a distance 
	// from the first shape point. The shape_dist_traveled field represents a real distance traveled along the route in units such as feet 
	// or kilometers. For example, if a bus travels a distance of 5.25 kilometers from the start of the shape to the stop, the shape_dist_traveled 
	// for the stop ID would be entered as "5.25". This information allows the trip planner to determine how much of the shape to draw when showing 
	// part of a trip on the map. The values used for shape_dist_traveled must increase along with stop_sequence: they cannot be used to show 
	// reverse travel along a route.
	// The units used for shape_dist_traveled in the stop_times.txt file must match the units that are used for this field in the shapes.txt file.
	ShapeDistTraveled float64

	feed *Feed
}

func (st *StopTime) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "trip_id":
		st.Trip = st.feed.Trips[val]
		break
	case "arrival_time":
		v, err := timeOfDayStringToSeconds(val)
		if err != nil {
			panic(err.Error() + val)
		}
		st.ArrivalTime = v
		break
	case "departure_time":
		v, err := timeOfDayStringToSeconds(val)
		if err != nil {
			panic(err)
		}
		st.DepartureTime = v
		break
	case "stop_id":
		st.Stop = st.feed.StopCollection.Stops[val]
		break
	case "stop_sequence":
		v, _ := strconv.Atoi(val)
		st.StopSequence = uint(v)
		break
	case "stop_headsign":
		st.Headsign = val
		break
	case "pickup_type":
		v, _ := strconv.Atoi(val) // Should panic on error !
		if v == 0 {
			st.PickupType = PickupRegular
		} else if v == 1 {
			st.PickupType = PickupUnavailable
		} else if v == 2 {
			st.PickupType = PickupThePhone
		} else if v == 3 {
			st.PickupType = PickupTheDriver
		}
		break
	case "drop_off_type":
		v, _ := strconv.Atoi(val) // Should panic on error !
		if v == 0 {
			st.DropOffType = DropOffRegular
		} else if v == 1 {
			st.DropOffType = DropOffUnavailable
		} else if v == 2 {
			st.DropOffType = DropOffThePhone
		} else if v == 3 {
			st.DropOffType = DropOffTheDriver
		}
		break
	case "shape_dist_traveled":
		v, _ := strconv.ParseFloat(val, 64) // Should panic on error !
		st.ShapeDistTraveled = v
		break

		// 
		// drop_off_type
		// shape_dist_traveled
	}
}

func timeOfDayStringToSeconds(t string) (uint, error) {
	components := strings.SplitN(t, ":", 3)
	hours, err := strconv.Atoi(components[0])
	if err != nil {
		hours, err = strconv.Atoi(string(components[0][1]))
		if err != nil {
			return 0, err
		}
	}
	minutes, err := strconv.Atoi(components[1])
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.Atoi(components[2])
	if err != nil {
		return 0, err
	}
	return uint((hours * 60 * 60) + (minutes * 60) + seconds), nil
}

func timeOfDayInSeconds(t *time.Time) uint {
	return uint(t.Hour()*60*60 + t.Minute()*60 + t.Second())
}
