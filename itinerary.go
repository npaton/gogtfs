package gtfs

import (
	"fmt"
	"log"
	"time"
)

type Itinerary struct {
	From *Stop
	To   *Stop

	MaxTransfers            uint // Default: 3
	MaxDuration             uint // In seconds. Default: 60*60*3 (3 hours)
	MaxWaitDuration         uint // In seconds. Default: 60*15 (15 min)
	DefaultTransferDuration uint // In seconds. Default: 60*5 (5 min)

	Departure *time.Time
	Arrival   *time.Time

	feed           *Feed
	trips          []*Trip
	bestTimeAtStop map[*Stop]int

	departureTime uint
	arrivalTime   uint
}

type Segment struct {
	From          *Stop
	To            *Stop
	DepartureTime *time.Time
	ArrivalTime   *time.Time
	Steps         []Step
}

func NewItinerary(f *Feed) (i *Itinerary) {
	i = &Itinerary{}
	i.MaxDuration = 60 * 60 * 3
	i.MaxTransfers = 3
	i.MaxWaitDuration = 60 * 60 * 15
	i.DefaultTransferDuration = 60 * 5
	// i.Step = Step{Steps:make([]Step,0)}
	i.feed = f
	return
}

type TmpTrace struct {
	msg  string
	next *TmpTrace
}

func (t *TmpTrace) trace() string {
	result := t.msg
	next := t.next
	for next != nil {
		result += "\n" + next.msg
		next = next.next
	}
	return result
}

func (step *Step) retrace(previousTrace *TmpTrace) (result *TmpTrace) {

	result = &TmpTrace{step.ToString(), previousTrace}

	previousStep := step.Previous
	if previousStep != nil {
		return previousStep.retrace(result)
	}

	return result
}

func (i *Itinerary) Run() {
	// i.trips = i.feed.TripsForDayAndDayRange(i.DepartureTime, &DayRange{uint(i.DepartureTime.Hour), 60*30*3})
	if i.Departure != nil {
		i.departureTime = uint(i.Departure.Hour()*60*60 + i.Departure.Minute()*60 + i.Departure.Second())
	}
	if i.Arrival != nil {
		i.arrivalTime = uint(i.Arrival.Hour()*60*60 + i.Arrival.Minute()*60 + i.Arrival.Second())
	}

	// start off the fun HERE
	walked := make(chan bool, 1)
	stepped := make(chan *Step, 10)
	found := make(chan *Step, 1)
	timeout := time.NewTicker(60 * 1 * 1E9).C

	var cheapestStep *Step
	var otherStep *Step
	var firstStep *Step
	for _, stoptime := range i.From.StopTimes {
		if stoptime.DepartureTime >= i.departureTime && stoptime.DepartureTime-i.departureTime < i.MaxWaitDuration {
			if firstStep == nil {
				firstStep = &Step{stoptime, nil, 0, 0, nil}
			} else {
				otherStep = &Step{stoptime, nil, 0, 0, otherStep}
				if cheapestStep == nil {
					cheapestStep = otherStep
				}
			}
		}
	}

	// log.Println("itinerary:",i.Departure, i.From.Name, i.To.Name, cheapestStep)

	i.Walk(cheapestStep, stepped, walked, found)
	// countWalked := 0
	// countStep := 0
	foundCount := 0
	for {
		select {
		case step := <-stepped:
			// log.Println("running step", step.ToString())
			if cheapestStep != nil && cheapestStep.Cost > step.Cost {
				step.parent = cheapestStep
				cheapestStep = step
			} else {
				currentStep := cheapestStep
				for {
					if currentStep.parent == nil || currentStep.parent.Cost < step.Cost {
						break
					}
					currentStep = currentStep.parent
				}
				step.parent = currentStep.parent
				currentStep.parent = step
			}

			// countStep+=1
			// if countStep % 1000 == 0 {
			// 	log.Println("countStep", countStep)
			// }

			break
		case <-walked:
			// log.Println("running walked")
			nextStep := cheapestStep
			if nextStep != nil {
				cheapestStep = cheapestStep.parent
				i.Walk(nextStep, stepped, walked, found)
			} else {
				panic("no routes found!!")
			}

			// countWalked+=1
			// if countWalked % 1000 == 0 {
			// 	log.Println("countWalked", countWalked)
			// }

			break
		case step := <-found:
			foundCount += 1
			// log.Println("running found", foundCount, ":", step.ToString())
			trace := step.retrace(nil)
			log.Println("===== Found")
			log.Println(trace.trace())
			if foundCount == 10 {
				return
			}
			break
			// panic("Found!!"+step.ToString())
		case <-timeout: // Timout after 10s
			panic("Itinerary more than 1 min to complete")
		}
	}
}

type Step struct {
	To       *StopTime // stop time index in trip
	Previous *Step
	Cost     int
	Changes  int

	// This is for the sort linked list
	parent *Step
}

func (s *Step) ToString() string {
	return fmt.Sprintf("%v - cost: %v - changes:%v", s.To.Stop.Name, s.Cost, s.Changes)
}

func (i *Itinerary) Walk(step *Step, stepped chan *Step, walked chan bool, found chan *Step) {
	go func() {
		if finalStopTime, cost, yes := step.To.Trip.RunsFromTo(step.To.Stop, i.To); yes {
			// if step.To.ArrivalTime+uint(step.Cost)+i.DefaultTransferDuration <= finalStopTime.DepartureTime && uint(step.Changes+1) <= i.MaxTransfers && uint(cost)+uint(step.Cost)+i.DefaultTransferDuration < i.MaxDuration {
			found <- &Step{
				To:       finalStopTime,
				Previous: step,
				Cost:     step.Cost + cost + int(i.DefaultTransferDuration),
				Changes:  step.Changes,
			}
			// }
		} else {
			nextStopTimeWithTransfer, cost, ok := step.To.Trip.NextStopTimeWithTransfer(step.To.Stop, nil)
			for ok {
				for _, st := range nextStopTimeWithTransfer.Stop.StopTimes {
					if st.Trip.Route != nextStopTimeWithTransfer.Trip.Route { // If we change route
						waitCost := st.DepartureTime - step.To.ArrivalTime
						if waitCost < i.MaxWaitDuration && nextStopTimeWithTransfer.ArrivalTime+i.DefaultTransferDuration <= st.DepartureTime && uint(step.Changes+1) <= i.MaxTransfers && waitCost+uint(cost)+uint(step.Cost)+i.DefaultTransferDuration < i.MaxDuration {
							stepped <- &Step{
								To:       st,
								Previous: step,
								Cost:     int(st.DepartureTime-step.To.ArrivalTime) + step.Cost + cost + int(i.DefaultTransferDuration),
								Changes:  step.Changes + 1,
							}
						}
					}
				}
				nextStopTimeWithTransfer, cost, ok = nextStopTimeWithTransfer.Trip.NextStopTimeWithTransfer(step.To.Stop, nextStopTimeWithTransfer.Stop)
			}
		}

		walked <- true
	}()
}

func tripsIntersectingStop(trips []*Trip, stop *Stop) []*Trip {
	foundTrips := make([]*Trip, 0, len(trips))
	for _, trip := range trips {
		if trip.RunsAccross(stop) {
			foundTrips = append(foundTrips, trip)
		}
	}
	return foundTrips
}
