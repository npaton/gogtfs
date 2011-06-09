include $(GOROOT)/src/Make.inc

TARG=gtfs
GOFILES=\
	gtfs.go\
	feed.go\
	parser.go\
	agency.go\
	stop.go\
	route.go\
	service.go\
	block.go\
	trip.go\
	stoptime.go\
	calendar.go\
	calendardate.go\
	fareattribute.go\
	farerule.go\
	shape.go\
	frequency.go\
	transfer.go\
	zone.go\

include $(GOROOT)/src/Make.pkg

