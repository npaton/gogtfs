package gtfs

import (
	"strconv"
	"time"
)

// Stop.LocationType possible values:
const (
	LocationTypeStop    = iota // 0 - Stop. A location where passengers board or disembark from a transit vehicle.
	LocationTypeStation        // 1 - Station. A physical structure or area that contains one or more stop.
)


// stops.txt
type Stop struct {

	// stop_id - Required. The stop_id field contains an ID that uniquely identifies a stop or station. 
	// Multiple routes may use the same stop. The stop_id is dataset unique.
	Id string

	// stop_code - Optional. The stop_code field contains short text or a number that uniquely identifies 
	// the stop for passengers. Stop codes are often used in phone-based transit information systems or printed 
	// on stop signage to make it easier for riders to get a stop schedule or real-time arrival information 
	// for a particular stop.
	// The stop_code field should only be used for stop codes that are displayed to passengers. 
	// For internal codes, use stop_id. This field should be left blank for stops without a code.
	Code string

	// stop_name - Required. The stop_name field contains the name of a stop or station. 
	// Please use a name that people will understand in the local and tourist vernacular.
	Name string

	// stop_desc - Optional. The stop_desc field contains a description of a stop. 
	// Please provide useful, quality information. Do not simply duplicate the name of the stop.
	Desc string

	// stop_lat - Required. The stop_lat field contains the latitude of a stop or station. 
	// The field value must be a valid WGS 84 latitude.
	Lat float64

	// stop_lon - Required. The stop_lon field contains the longitude of a stop or station. 
	// The field value must be a valid WGS 84 longitude value from -180 to 180.
	Lon float64

	// zone_id - Optional. The zone_id field defines the fare zone for a stop ID. 
	// Zone IDs are required if you want to provide fare information using fare_rules.txt. 
	// If this stop ID represents a station, the zone ID is ignored.
	ZoneId string

	// stop_url - Optional. The stop_url field contains the URL of a web page about a particular stop. 
	// This should be different from the agency_url and the route_url fields. 
	// The value must be a fully qualified URL that includes http:// or https://, and any special characters 
	// in the URL must be correctly escaped. See http://www.w3.org/Addressing/URL/4_URI_Recommentations.html 
	// for a description of how to create fully qualified URL values.
	Url string

	// location_type - ptional. The location_type field identifies whether this stop ID represents a stop or station. 
	// If no location type is specified, or the location_type is blank, stop IDs are treated as stops. Stations may 
	// have different properties from stops when they are represented on a map or used in trip planning.
	// The location type field can have the following values:
	// 		0 or blank - Stop. A location where passengers board or disembark from a transit vehicle.
	// 		1 - Station. A physical structure or area that contains one or more stop.
	LocationType byte

	// parent_station - Optional. For stops that are physically located inside stations, the parent_station field 
	// identifies the station associated with the stop. To use this field, stops.txt must also contain a row where 
	// this stop ID is assigned location type=1.
	// 	 This stop ID represents...      	This entry's location type...	This entry's parent_station field contains...
	// 	 A stop located inside a station.	0 or blank                   	The stop ID of the station where this stop is located. 
	// 	                                 	                             	The stop referenced by parent_station must have location_type=1.
	// 	 A stop located outside a station.	0 or blank                   	A blank value. The parent_station field doesn't apply to this stop.
	// 	 A station.                       	1                            	A blank value. Stations can't contain other stations.
	ParentStationId string
	
	
	Transfers map[string]*Transfer
	
	StopTimes []*StopTime

	feed *Feed
}

func NewStop() *Stop {
	return &Stop{Transfers:make(map[string]*Transfer)}
}

func (s *Stop) ParentStation() *Stop {
	return s.feed.StopCollection.Stops[s.ParentStationId]
}

func (s *Stop) NextStopTimes(time *time.Time, count int) (stopTimes []*StopTime) {
	timeOfDay := timeOfDayInSeconds(time)
	for _, stoptime := range s.StopTimes {
		if stoptime.Trip.RunsOn(time) && stoptime.DepartureTime > timeOfDay {
			stopTimes = append(stopTimes, stoptime)
		}
	}
	return
}


func (s *Stop) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "stop_id":
		s.Id = val
		break
	case "stop_code":
		s.Code = val
		break
	case "stop_name":
		s.Name = val
		break
	case "stop_desc":
		s.Desc = val
		break
	case "stop_lat":
		v, _ := strconv.Atof64(val) // Should panic on error !
		s.Lat = v
		break
	case "stop_lon":
		v, _ := strconv.Atof64(val) // Should panic on error !
		s.Lon = v
		break
	case "zone_id":
		s.ZoneId = val
		break
	case "stop_url":
		s.Url = val
		break
	case "location_type":
		v, _ := strconv.Atoui(val) // Should panic on error !
		if v == 0 {
			s.LocationType = LocationTypeStop
		} else if v == 1 {
			s.LocationType = LocationTypeStation
		}
		break
	case "parent_station":
		s.ParentStationId = val
		break
	}
}
