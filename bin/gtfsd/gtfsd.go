package main

// import _ "http/pprof"

import (
	"log"
	"gtfs"
	"flag"
	"os"
	"strings"
	"path/filepath"
	"http"
	"fmt"
	"json"
	"time"
	"websocket"
)

var (
	pathsString = flag.String("d", "", "Directories containing gtfs txt or zip files or zip file path (directories are traversed, multi coma separated: \"/here,/there\")")
	shouldLog   = flag.Bool("v", false, "Log to Stdout/err")
	needHelp    = flag.Bool("h", false, "Displays this help message...")
	listenAddr  = flag.String("http", ":8080", "HTTP listen address")
	prefPath    string
	homeDir     string
	feeds       map[string]*gtfs.Feed
)

func init() {
	homeDir = os.Getenv("HOME")
	flag.StringVar(&prefPath, "p", homeDir+"/.gtfs", "Preference file path")
	feeds = make(map[string]*gtfs.Feed, 10)
}


type Preferences struct {
	Paths []string
}

func (p *Preferences) saveToDisk() {
}


func discoverGtfsPaths(path string) (results []string) {
	// log.Println("discoverGtfsPaths")
	path = filepath.Clean(path)
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return
	}

	if fileInfo.IsDirectory() {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return
		}

		requiredFiles := gtfs.RequiredFiles
		requiredCalendarFiles := gtfs.RequiredEitherCalendarFiles
		foundCalendar := false
		foundFiles := make([]string, 0, len(requiredFiles))
		for _, fi := range fileInfos {
			if fi.IsDirectory() {
				subdirectoryResults := discoverGtfsPaths(path + "/" + fi.Name)
				for _, newpath := range subdirectoryResults {
					results = append(results, newpath)
				}
			} else if filepath.Ext(fi.Name) == ".zip" {
				results = append(results, path+"/"+fi.Name)
			} else {
				for _, f := range requiredFiles {
					if fi.Name == f { // This loops a little too much but hey...
						foundFiles = append(foundFiles, f)
					}
				}
				if !foundCalendar {
					for _, f := range requiredCalendarFiles {
						if fi.Name == f {
							foundCalendar = true
						}
					}
				}
				
			}
		}
		
		if len(foundFiles) == len(requiredFiles) && foundCalendar {
			results = append(results, path)
		}
		
		
	} else {
		if filepath.Ext(path) == ".zip" {
			results = append(results, path)
		}
	}
	return
}

func main() {
	flag.Parse()

	if *needHelp {
		flag.Usage()
		os.Exit(0)
	}

	pathsAll := strings.Split(*pathsString, ",", -1)
	paths := make([]string, 0, len(pathsAll))

	for _, path := range pathsAll {
		if path != "" {

			// Why do I have to do this
			path = strings.Replace(path, "~", homeDir, 1)

			newpaths := discoverGtfsPaths(path)
			// log.Println("Paths", len(newpaths), newpaths)

			if len(newpaths) > 0 {
				for _, p := range newpaths {
					paths = append(paths, p)
				}
			}
		}
	}

	log.Println("Paths", paths)

	if !*shouldLog {
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0) // Shouldn't be an error
		defer devNull.Close()                                 // Useless, is it not?
		log.SetOutput(devNull)
	}

	log.SetPrefix("gtfs - ")

	channels := make([]chan bool, 0, len(paths))
	totalStopTimes := 0
	if len(paths) > 0 {
		for _, path := range paths[:] {
			if path != "" {
				channel := make(chan bool)
				channels = append(channels, channel)
				go func(path string, ch chan bool) {
					log.Println("Started loading", path)
					feed, err := gtfs.NewFeed(path)
					if err != nil {
						log.Fatal(err)
					} else {
						feed.Load()
						feeds[path] = feed
						currentFeed = path
					}
					totalStopTimes = totalStopTimes + feed.StopTimesCount
					channel <- true
				}(path, channel)
			}
		}
	}

	if *listenAddr != "-" {
		go func() {
			// Waiting for jobs to finnish
			for _, c := range channels {
				<-c
			}

			log.Println("Total StopTimes count", totalStopTimes)
			log.Println("Finished")
		}()

		http.HandleFunc("/", Index)
		http.HandleFunc("/shapes.json", Shapes)
		http.HandleFunc("/trips.json", Trips)
		http.Handle("/gtfs", websocket.Handler(GTFSSocket))
		http.HandleFunc("/load", LoadFeed)

		httpInitErr := http.ListenAndServe(*listenAddr, nil)
		if httpInitErr != nil {
			log.Printf("HTTP init error: %v\n", httpInitErr)
		}

	} else {
		// Waiting for jobs to finnish
		for _, c := range channels {
			<-c
		}
	}
}

var currentFeed string

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}


func Shapes(w http.ResponseWriter, r *http.Request) {
	log.Println("Shapes")

	if currentFeed == "" {
		fmt.Fprint(w, "{}")
		return
	}
	shapes, err := json.Marshal(feeds[currentFeed].Shapes)

	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(shapes))
	log.Println("Shapes done")
	return

}

func Trips(w http.ResponseWriter, r *http.Request) {
	log.Println("Trips")

	if currentFeed == "" {
		fmt.Fprint(w, "{}")
		return
	}

	ts := feeds[currentFeed].TripsForDay(time.LocalTime())
	trips, err := json.Marshal(ts)

	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(trips))
	log.Println("Trips done")
	return
}

func LoadFeed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Form Parse Error!?", err)
		return
	}

	if feeds[r.Form["feed"][0]] == nil {
		fmt.Fprintf(w, "{\"error\":\"Error loading %v (not found)\"}", r.Form["feed"][0])
		return
	}

	feeds[r.Form["feed"][0]].Load()
	fmt.Fprintf(w, "{\"success\":\"Loading %v...\"}", r.Form["feed"][0])

	return // http.Redirect(w, r, r.Referer, http.StatusTemporaryRedirect) // Come back
}


type WebSocketContext struct {
	ws          *websocket.Conn
	inMessages  chan string
	outMessages chan string
	closeCh     chan bool
}

// Echo the data received on the Web Socket.
func GTFSSocket(ws *websocket.Conn) {
	ctx := WebSocketContext{ws, make(chan string), make(chan string), make(chan bool)}
	ctx.Handle()
}

func (ctx *WebSocketContext) Handle() {

	go func() {
		for {
			select {
			case in := <-ctx.inMessages:
				log.Println("In Message", in, len(in))
			case out := <-ctx.outMessages:
				outb := []uint8(out)
				// log.Println("Out Message", outb, "ASSTRING:", out)
				if _, err := ctx.ws.Write(outb); err != nil {
					log.Println("W", err)
					ctx.closeCh <- true
				}
			case <-ctx.closeCh:
				log.Println("Closed connection")
				ctx.ws.Close()
				break
			}
		}
	}()

	go func() {
		time.Sleep(5 * 1E9)
		ctx.outMessages <- "Hello"
		// time.Sleep(1*1E9)
		// ctx.outMessages <- "How are you"
		// time.Sleep(1*1E9)
		// ctx.outMessages <- "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	}()

	for {
		buf := make([]byte, 256)
		n, err := ctx.ws.Read(buf)
		if err != nil {
			log.Println(err)
			ctx.closeCh <- true
			break
		} else {
			ctx.inMessages <- string(buf[0:n])
		}
	}
}
