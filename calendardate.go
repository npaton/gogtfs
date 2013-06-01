package gtfs

import (
	// "time"
	// "log"
	"strconv"
)

// CalendarDate.ExceptionType possible values:
const (
	_                               = iota // Ignore 0
	CalendarExceptionAddedService          // A value of 1 indicates that service has been added for the specified date.
	CalendarExceptionRemovedService        // A value of 2 indicates that service has been removed for the specified date.
)

// calendar_dates.txt
// This file is optional. The calendar_dates table allows you to explicitly activate or disable service IDs by date. You can use it in two ways.
// Recommended: Use calendar_dates.txt in conjunction with calendar.txt, where calendar_dates.txt defines any exceptions to the default service 
// categories defined in the calendar.txt file. If your service is generally regular, with a few changes on explicit dates (for example, to 
// accomodate special event services, or a school schedule), this is a good approach.
// Alternate: Omit calendar.txt, and include ALL dates of service in calendar_dates.txt. If your schedule varies most days of the month, or you 
// want to programmatically output service dates without specifying a normal weekly schedule, this approach may be preferable.
type CalendarDate struct {

	// service_id - Required. The service_id contains an ID that uniquely identifies a set of dates when a service 
	// exception is available for one or more routes. Each (service_id, date) pair can only appear once in calendar_dates.txt.
	// If the a service_id value appears in both the calendar.txt and calendar_dates.txt files, the information in 
	// calendar_dates.txt modifies the service information specified in calendar.txt. This field is referenced by the trips.txt file.
	serviceId string

	// date - Required. The date field specifies a particular date when service availability is different than the norm. 
	// You can use the exception_type field to indicate whether service is available on the specified date.
	// The date field's value should be in YYYYMMDD format.
	Date int

	// exception_type - Required. The exception_type indicates whether service is available on the date specified in the date field.
	// 	 A value of 1 indicates that service has been added for the specified date.
	// 	 A value of 2 indicates that service has been removed for the specified date.
	// For example, suppose a route has one set of trips available on holidays and another set of trips available on all other days. 
	// You could have one service_id that corresponds to the regular service schedule and another service_id that corresponds to the 
	// holiday schedule. For a particular holiday, you would use the calendar_dates file to add the holiday to the holiday service_id 
	// and to remove the holiday from the regular service_id schedule.
	// See CalendarException constants
	ExceptionType byte

	feed *Feed
}

// func (cd *CalendarDate) ExceptionOn(date string) (exceptionalDate, shouldRun bool) {
// 	exceptionalDate, shouldRun = false, false
// 	if stringDayDateComp(cd.Date, date) == 0 {
func (cd *CalendarDate) ExceptionOn(intday int) (exceptionalDate, shouldRun bool) {
	exceptionalDate, shouldRun = false, false
	if cd.Date == intday {
		exceptionalDate = true
		if cd.ExceptionType == CalendarExceptionAddedService {
			shouldRun = true
		}
	}
	return exceptionalDate, shouldRun
}

func (cd *CalendarDate) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "service_id":
		cd.serviceId = val
		break
	case "date":
		v, _ := strconv.Atoi(val) // Should panic on error !
		cd.Date = v
		break
	// case "date":
	// 	cd.Date = val
	// 	break
	case "exception_type":
		if val == "1" {
			cd.ExceptionType = 1
		} else {
			cd.ExceptionType = 0
		}
		break
	}
}
