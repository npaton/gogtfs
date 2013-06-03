package gtfs

import (
	"fmt"
	"strconv"
	"time"
	// "log"
)

// calendar.txt
type Calendar struct {

	// service_id - Required. The service_id contains an ID that uniquely identifies a set of dates when service is 
	// available for one or more routes. Each service_id value can appear at most once in a calendar.txt file. This
	// value is dataset unique. It is referenced by the trips.txt file.
	serviceId string

	// monday - Required. The monday field contains a binary value that indicates whether the service is valid for all Mondays.
	// 	 A value of 1 indicates that service is available for all Mondays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Mondays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Monday bool

	// tuesday	- Required. The tuesday field contains a binary value that indicates whether the service is valid for all Tuesdays.
	// 	 A value of 1 indicates that service is available for all Tuesdays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Tuesdays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Tuesday bool

	// wednesday - Required. The wednesday field contains a binary value that indicates whether the service is valid for all Wednesdays.
	// 	 A value of 1 indicates that service is available for all Wednesdays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Wednesdays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Wednesday bool

	// thursday - Required. The thursday field contains a binary value that indicates whether the service is valid for all Thursdays.
	// 	 A value of 1 indicates that service is available for all Thursdays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Thursdays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Thursday bool

	// friday - Required. The friday field contains a binary value that indicates whether the service is valid for all Fridays.
	// 	 A value of 1 indicates that service is available for all Fridays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Fridays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file
	Friday bool

	// saturday - Required. The saturday field contains a binary value that indicates whether the service is valid for all Saturdays.
	// 	 A value of 1 indicates that service is available for all Saturdays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Saturdays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Saturday bool

	// sunday - Required. The sunday field contains a binary value that indicates whether the service is valid for all Sundays.
	// 	 A value of 1 indicates that service is available for all Sundays in the date range. (The date range is specified using the start_date and end_date fields.)
	// 	 A value of 0 indicates that service is not available on Sundays in the date range.
	// Note: You may list exceptions for particular dates, such as holidays, in the calendar_dates.txt file.
	Sunday bool

	// start_date - Required. The start_date field contains the start date for the service.
	// The start_date field's value should be in YYYYMMDD format.
	StartDate int

	// end_date - Required. The end_date field contains the end date for the service. This date is included in the service interval.
	// The end_date field's value should be in YYYYMMDD format.
	EndDate int

	feed *Feed
}

// func (c *Calendar) ValidOn(date string) bool {
// 	if stringDayDateComp(c.StartDate, date) <= 0 && stringDayDateComp(c.EndDate, date) >= 0 {
func (c *Calendar) ValidOn(intday int, t *time.Time) bool {
	if c.StartDate <= intday && c.EndDate >= intday {

		switch t.Weekday() {
		case time.Monday:
			return c.Monday
		case time.Tuesday:
			return c.Tuesday
		case time.Wednesday:
			return c.Wednesday
		case time.Thursday:
			return c.Thursday
		case time.Friday:
			return c.Friday
		case time.Saturday:
			return c.Saturday
		case time.Sunday:
			return c.Sunday
		default:
			return false
		}

	}

	return false
}

func (c *Calendar) setField(fieldName, val string) {
	// log.Println("setField", fieldName, value)
	switch fieldName {
	case "service_id":
		c.serviceId = val
		break
	case "monday":
		if val == "1" {
			c.Monday = true
		} else {
			c.Monday = false
		}
		break
	case "tuesday":
		if val == "1" {
			c.Tuesday = true
		} else {
			c.Tuesday = false
		}
		break
	case "wednesday":
		if val == "1" {
			c.Wednesday = true
		} else {
			c.Wednesday = false
		}
		break
	case "thursday":
		if val == "1" {
			c.Thursday = true
		} else {
			c.Thursday = false
		}
		break
	case "friday":
		if val == "1" {
			c.Friday = true
		} else {
			c.Friday = false
		}
		break
	case "saturday":
		if val == "1" {
			c.Saturday = true
		} else {
			c.Saturday = false
		}
		break
	case "sunday":
		if val == "1" {
			c.Sunday = true
		} else {
			c.Sunday = false
		}
		break
	case "start_date":
		v, _ := strconv.Atoi(val) // Should panic on error !
		c.StartDate = v
		break
	case "end_date":
		v, _ := strconv.Atoi(val) // Should panic on error !
		c.EndDate = v
		break
		// case "start_date":
		// 	c.StartDate = val
		// 	break
		// case "end_date":
		// 	c.EndDate = val
		// 	break
	}
}

func TimeToStringDate(time *time.Time) string {
	return time.Format("20060102")
}

func StringDateToTime(date string) (time.Time, error) {
	return time.Parse("20060102", date) // Time parsing almost sucks
}

func stringDayDateComp(dateA, dateB string) int {
	if dateA == dateB {
		return 0
	}

	if dateA[0:4] == dateB[0:4] {
		if dateA[4:6] == dateB[4:6] {
			vA, errA := strconv.Atoi(dateA[6:8])
			vB, errB := strconv.Atoi(dateB[6:8])
			// log.Println("Compare days", vA, vB, "-", dateA[6:8], "-", string(dateB[6:8]), "-", dateB[0:4], "-", string(dateB[0:4]), "-")
			if errA != nil && errB != nil {
				panic(fmt.Sprintln("stringDayDateComp impossible to comparable dates:", dateA, dateB))
			}
			if vA > vB {
				return 1
			} else {
				return -1
			}
		}

		vA, errA := strconv.Atoi(dateA[4:6])
		vB, errB := strconv.Atoi(dateB[4:6])
		// log.Println("Compare months", vA, vB)
		if errA != nil && errB != nil {
			panic(fmt.Sprintln("stringDayDateComp impossible to comparable dates:", dateA, dateB))
		}
		if vA > vB {
			return 1
		} else {
			return -1
		}
	}

	vA, errA := strconv.Atoi(dateA[0:4])
	vB, errB := strconv.Atoi(dateB[0:4])
	// log.Println("Compare years", vA, vB)
	if errA != nil && errB != nil {
		panic(fmt.Sprintln("stringDayDateComp impossible to comparable dates:", dateA, dateB))
	}
	if vA > vB {
		return 1
	} else {
		return -1
	}

	return 0
}
