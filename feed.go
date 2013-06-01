package gtfs

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	// "runtime"
	// "bufio"
	"time"

	// "fmt"
	// "strings"
)

type Feed struct {
	loadedDate time.Time
	path       string // Feed's path on disk (zip or folder containing GTFS .txt files)
	Agencies   map[string]*Agency
	// Stops          map[string]*Stop
	StopCollection
	Routes           map[string]*Route
	Trips            map[string]*Trip
	Services         map[string]*Service
	Shapes           map[string]*Shape
	Calendars        map[string]*Calendar
	CalendarDates    map[string][]*CalendarDate
	Loaded           bool
	StopTimesCount   int
	TranfersCount    int
	FrequenciesCount int
}

var RequiredFiles = []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt"}
var RequiredEitherCalendarFiles = []string{"calendar.txt", "calendar_dates.txt"}
var AllFiles = []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt", "calendar.txt", "calendar_dates.txt", "fare_attributes.txt", "fare_rules.txt", "shapes.txt", "frequencies.txt", "transfers.txt"}

func NewFeed(path string) (*Feed, error) {

	feed := &Feed{
		path:           path,
		Agencies:       make(map[string]*Agency),
		StopCollection: NewStopCollection(),
		Routes:         make(map[string]*Route),
		Trips:          make(map[string]*Trip),
		Services:       make(map[string]*Service),
		Shapes:         make(map[string]*Shape),
		Calendars:      make(map[string]*Calendar),
		CalendarDates:  make(map[string][]*CalendarDate),
		Loaded:         false,
	}

	return feed, nil
}
func (f *Feed) Reload() error {
	f.Agencies = make(map[string]*Agency)
	f.StopCollection = NewStopCollection()
	f.Routes = make(map[string]*Route)
	f.Trips = make(map[string]*Trip)
	f.Services = make(map[string]*Service)
	f.Shapes = make(map[string]*Shape)
	f.Calendars = make(map[string]*Calendar)
	f.CalendarDates = make(map[string][]*CalendarDate)
	f.Loaded = false
	return f.Load()
}

func (f *Feed) Load() error {
	if f.Loaded {
		return nil
	}

	if filepath.Ext(f.path) == ".zip" {

		zipReader, err := zip.OpenReader(f.path)
		if err != nil {
			return err
		}

		if zipReader.Comment != "" {
			log.Println("zipReader.Comment", zipReader.Comment)
		}

		// Sort for loading dependencies
		fileIndexes := make([]int, len(zipReader.File))
		for _, f := range AllFiles[:] {
			for i, zf := range zipReader.File[:] {
				if zf.FileHeader.Name == f {
					fileIndexes = append(fileIndexes, i)
				}
			}
		}

		// Open and parse files
		for _, fileIndex := range fileIndexes {
			reader, err := zipReader.File[fileIndex].Open()
			if err != nil {
				log.Println(err)
				err = nil
			}

			err = f.parseTxtFile(reader, zipReader.File[fileIndex].FileHeader.Name)
			if err != nil {
				log.Println(err)
			}
		}

	} else {
		for i, fileName := range AllFiles[:] {
			err := f.openAndParseTxtFile(f.path, fileName)
			if err != nil {
				if i < len(RequiredFiles) {
					log.Println("Error in file", f.path+"/"+fileName, err)
				}
			}
		}
	}

	// Color field copy from Routes to Shapes for json export
	// And calculate the DayRange for each trip
	bench("Trips calculations", func() interface{} {
		for _, trip := range f.Trips {
			trip.copyColorToShape()
			trip.calculateDayTimeRange()
			for _, freq := range trip.Frequencies {
				freq.calculateDayTimeRange()
			}
		}
		return "yes"
	})

	if len(f.Agencies) == 0 {
		return errors.New("A feed needs a least one agency !")
	}

	log.Println("Agency count", len(f.Agencies))
	log.Println("Stops count", f.StopCollection.Length())
	log.Println("Routes count", len(f.Routes))
	log.Println("Trip count", len(f.Trips))
	log.Println("StopTimes count", f.StopTimesCount)
	log.Println("Shapes count", len(f.Shapes))
	log.Println("Calendars count", len(f.Calendars))
	log.Println("CalendarDates count", len(f.CalendarDates))
	log.Println("Tranfers count", f.TranfersCount)
	log.Println("FrequenciesCount count", f.TranfersCount)

	// log.Printf("gtfsd weight - bytes = %d - footprint = %d", runtime.MemStats.HeapAlloc, runtime.MemStats.Sys)
	// now := time.Now().Local()
	// tripsForDay := f.TripsForDay(&now)
	// bench("Trip today count", func() interface{} {
	// 	return len(tripsForDay)
	// })
	// bench("Trip between 6 and 7 today count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRange(&now, &DayRange{60 * 60 * 6, 60 * 60}))
	// })
	// bench("Trip between 6 and 8 today count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRange(&now, &DayRange{60 * 60 * 6, 60 * 60 * 2}))
	// })
	// bench("Trip between 7 and 8 today count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRange(&now, &DayRange{60 * 60 * 7, 60 * 60}))
	// })
	// bench("Trip between 15 and 20 today count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRange(&now, &DayRange{60 * 60 * 15, 60 * 60 * 5}))
	// })
	// bench("Trip between 17 and 19 today count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRange(&now, &DayRange{60 * 60 * 17, 60 * 60 * 2}))
	// })

	// stopX := f.StopCollection.RandomStop()
	// bench("Trip between 19:00 and 19:10 today on stop x count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRangeAndStop(&now, &DayRange{60 * 60 * 0, 60 * 10}, stopX))
	// })

	// // for {
	// // 	trips := f.TripsForDayAndDayRangeAndStop(&now, &DayRange{60*60*0, 60*10}, stopX)
	// // 	log.Println("found trips", trips)
	// // }

	// bench("Trip all day today on stop x count", func() interface{} {
	// 	return len(f.TripsForDayAndDayRangeAndStop(&now, &DayRange{60 * 60 * 0, 60 * 60 * 24}, stopX))
	// })
	// var memstats runtime.MemStats
	// runtime.ReadMemStats(&memstats)
	// log.Printf("Before GC - bytes = %d - footprint = %d", memstats.HeapAlloc, memstats.Sys)
	// runtime.GC()
	// runtime.ReadMemStats(&memstats)
	// log.Printf("After GC - bytes = %d - footprint = %d", memstats.HeapAlloc, memstats.Sys)

	// stopY := f.StopCollection.RandomStop()
	// runtime.ReadMemStats(&memstats)
	// log.Printf("gtfsd weight before itinerary - bytes = %d - footprint = %d", memstats.HeapAlloc, memstats.Sys)
	// // for {
	// // 	i := NewItinerary(f)
	// // 	t := &now
	// // 	t.Hour = 7
	// // 	i.Departure = t
	// // 	i.From = stopX
	// // 	i.To = stopY
	// // 	i.Run()
	// // }
	// bench("Itinerary", func() interface{} {
	// 	i := NewItinerary(f)
	// 	now := time.Now().Local()
	// 	t := time.Date(now.Year(), now.Month(), now.Day(), 7, now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	// 	i.Departure = &t
	// 	i.From = stopX
	// 	i.To = stopY
	// 	i.Run()
	// 	return ""
	// })
	// runtime.ReadMemStats(&memstats)
	// log.Printf("gtfsd weight after itinerary - bytes = %d - footprint = %d", memstats.HeapAlloc, memstats.Sys)
	// runtime.GC()
	// runtime.ReadMemStats(&memstats)
	// log.Printf("After GC - bytes = %d - footprint = %d", memstats.HeapAlloc, memstats.Sys)

	return nil
}

func bench(name string, toBench func() interface{}) {
	start := time.Now()
	result := toBench()
	duration := time.Since(start)
	log.Println("Bench '"+name+"' took:", float64(duration.Nanoseconds())/1E6, "ms", "- result:", result)
}

func (feed *Feed) TripsForDay(date *time.Time) []*Trip {
	tripsos := make([]*Trip, 0, len(feed.Trips))
	for _, trip := range feed.Trips {
		if trip.RunsOn(date) {
			// log.Println("Trip on", date+":", trip)
			tripsos = append(tripsos, trip)
		}
	}

	// log.Println("Trips on", date+":", len(tripsos))
	return tripsos
}

func (feed *Feed) TripsForDayAndDayRange(date *time.Time, dayrange *DayRange) []*Trip {
	tripsos := make([]*Trip, 0, len(feed.Trips))
	for _, trip := range feed.Trips {
		if trip.RunsOn(date) && trip.Intersects(dayrange) {
			// log.Println("Trip on", date+":", trip)
			tripsos = append(tripsos, trip)
		}
	}

	// log.Println("Trips on", date+":", len(tripsos))
	return tripsos
}

func (feed *Feed) TripsForDayAndDayRangeAndStop(date *time.Time, dayrange *DayRange, stop *Stop) []*Trip {
	tripsos := make([]*Trip, 0, len(feed.Trips))
	for _, trip := range feed.Trips {
		if trip.RunsOn(date) && trip.Intersects(dayrange) && trip.RunsAccross(stop) {
			// log.Println("Trip on", date+":", trip)
			tripsos = append(tripsos, trip)
		}
	}

	// log.Println("Trips on", date+":", len(tripsos))
	return tripsos
}

func (feed *Feed) openAndParseTxtFile(basePath, fileName string) (err error) {
	fullpath, err := filepath.Abs(filepath.Join(basePath, fileName))
	if err != nil {
		return err
	}

	file, err := os.Open(fullpath)
	if err != nil {
		return err
	}

	// if fileName == "stop_times.txt" {
	// 	fileForLineCount, err := os.Open(fullpath)
	// 	r := bufio.NewReader(fileForLineCount)
	// 	count := 0
	// 	_, p, err := r.ReadLine()
	// 	for err == nil {
	// 		if !p {
	// 			count = count + 1
	// 		}
	// 		_, p, err = r.ReadLine()
	// 	}
	// 
	// 	feed.StopTimes = make([]*StopTime, count-1, count)
	// }

	return feed.parseTxtFile(file, fileName)
}

func (feed *Feed) parseTxtFile(reader io.Reader, fileName string) (err error) {

	parser := new(Parser)
	switch fileName {
	case "agency.txt":
		log.Println("agency.txt")
		err = parser.parse(reader, func(k, v []string) {
			agency := new(Agency)
			agency.feed = feed
			fieldsSetter(agency, k, v)
			// log.Println("  - agency:", agency)
			feed.Agencies[agency.Id] = agency
		})
		if err != nil {
			return
		}
		break
	case "stops.txt":
		log.Println("stops.txt")
		err = parser.parse(reader, func(k, v []string) {
			stop := NewStop()
			stop.feed = feed
			fieldsSetter(stop, k, v)
			// log.Println("  - stop:", stop)
			feed.StopCollection.SetStop(stop.Id, stop)
		})
		if err != nil {
			return
		}
		break
	case "routes.txt":
		log.Println("routes.txt")
		err = parser.parse(reader, func(k, v []string) {
			route := new(Route)
			route.feed = feed
			fieldsSetter(route, k, v)
			// log.Println("  - route:", route)
			feed.Routes[route.Id] = route
		})
		if err != nil {
			return
		}
		break
	case "trips.txt":
		log.Println("trips.txt")
		err = parser.parse(reader, func(k, v []string) {
			trip := new(Trip)
			trip.feed = feed
			fieldsSetter(trip, k, v)
			trip.afterInit()
			// log.Println("  - trip:", trip)
			feed.Trips[trip.Id] = trip
		})
		if err != nil {
			return
		}
		break
	case "stop_times.txt":
		// break
		log.Println("stop_times.txt")
		err = parser.parse(reader, func(k, v []string) {
			stopTime := new(StopTime)
			stopTime.feed = feed
			fieldsSetter(stopTime, k, v)
			// log.Println("  - stopTime:", stopTime)
			if stopTime.Trip != nil {
				feed.StopTimesCount = feed.StopTimesCount + 1
				stopTime.Trip.AddStopTime(stopTime)
			}

			if stopTime.Stop != nil { // For security, should proceed otherwise
				stopTime.Stop.StopTimes = append(stopTime.Stop.StopTimes, stopTime)
			}

			// feed.StopTimes = append(feed.StopTimes, stopTime)
		})
		if err != nil {
			return
		}
		break

	case "calendar.txt":
		log.Println("calendar.txt")
		err = parser.parse(reader, func(k, v []string) {
			calendar := new(Calendar)
			calendar.feed = feed
			fieldsSetter(calendar, k, v)
			// log.Println("  - calendar:", calendar)
			feed.Calendars[calendar.serviceId] = calendar
		})
		if err != nil {
			return
		}
		break
	case "calendar_dates.txt":
		log.Println("calendar_dates.txt")
		err = parser.parse(reader, func(k, v []string) {
			calendardate := new(CalendarDate)
			calendardate.feed = feed
			fieldsSetter(calendardate, k, v)
			// log.Println("  - calendardate:", calendardate)
			if feed.CalendarDates[calendardate.serviceId] == nil {
				feed.CalendarDates[calendardate.serviceId] = make([]*CalendarDate, 0)
			}
			feed.CalendarDates[calendardate.serviceId] = append(feed.CalendarDates[calendardate.serviceId], calendardate)
		})
		if err != nil {
			return
		}
		break
	// case "fare_attributes.txt":
	// case "fare_rules.txt":
	case "shapes.txt":
		// break
		log.Println("shapes.txt")
		err = parser.parse(reader, func(k, v []string) {
			shapepoint := new(ShapePoint)
			shapepoint.feed = feed
			fieldsSetter(shapepoint, k, v)
			// log.Println("  - shapepoint:", shapepoint)
			if feed.Shapes[shapepoint.Id] == nil {
				feed.Shapes[shapepoint.Id] = new(Shape)
				feed.Shapes[shapepoint.Id].Points = []*ShapePoint{shapepoint}
				feed.Shapes[shapepoint.Id].Id = shapepoint.Id
			} else {
				feed.Shapes[shapepoint.Id].Points = append(feed.Shapes[shapepoint.Id].Points, shapepoint)
			}
		})
		if err != nil {
			return
		}
		break
	case "frequencies.txt":
		log.Println("frequencies.txt")
		err = parser.parse(reader, func(k, v []string) {
			frequency := new(Frequency)
			frequency.feed = feed
			fieldsSetter(frequency, k, v)
			// log.Println("  - frequency:", frequency)
			if frequency.Trip != nil {
				feed.Trips[frequency.Trip.Id].Frequencies = append(feed.Trips[frequency.Trip.Id].Frequencies, *frequency)
				feed.FrequenciesCount = feed.FrequenciesCount + 1
			}
		})
		if err != nil {
			return
		}
		break
	case "transfers.txt":
		log.Println("transfers.txt")
		err = parser.parse(reader, func(k, v []string) {
			transfer := new(Transfer)
			transfer.feed = feed
			fieldsSetter(transfer, k, v)
			// log.Println("  - transfer:", transfer)
			feed.StopCollection.Stop(transfer.FromStopId).Transfers[transfer.ToStopId] = transfer
			feed.TranfersCount = feed.TranfersCount + 1
		})
		if err != nil {
			return
		}
		break
	}

	return
}
