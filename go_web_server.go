package main

import (
	"log"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"time"
	"crypto/sha256"
	"fmt"
	"sync"
)

// Create a hashEvent struct containing a time Duration that is set five seconds// from the time the hash was created and a byte array to hold the hash checksum
// when it can return the hash.

type hashEventData struct {
	createTime	time.Duration
	sha256sum	[32]byte 	
}
var hashEventMap	map[int]*hashEventData

// everything we need to create or add to the map
type hashEvents struct {
	keyCounter	int
//	hashEventMap	map[int]*hashEventData
}


// Create an incrementing counter for counting # of hashed passwords 
var keyCounter	int
var mux		sync.Mutex

func root(w http.ResponseWriter, r *http.Request) {

	// Check if the current request URL path exactly matches the expected
	// patterns mapped to handler functions.
	//
	// A special case for the "/" root path is done to see if the pattern
	// matches the Key Value Pair for querying the hash {hash, map id} which
	// will call the showHash handler.  If no match is found, we return a
	// 404 response to the client.

	if r.URL.Path == "/" {

		http.NotFound(w, r)
		return
	}
}

/*
 * This serveMux approach uses strict pattern matches from the URL, 
 * which limits the patterns to strings.  A workaround for the /hash/id 
 * lookup limitation is to match everything at the "/" root, then search 
 * the body for the key value pair we want to map to the showStats function.
 *
 * It is optional to test with curl using -X POST or -X GET; it doesn't
 * matter, as the server figures out what the request is doing even when
 * those are not on the curl command line.
 */

/* 
 * If there is no valid input (hash, stats, shutdown), do nothing
 * and return a 404 error.
 */
func checkRootUrl(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL.Path = %s\n", r.URL.Path) 
	if r.URL.Path == "/" { 
		http.NotFound(w, r)
	}
	return
}

// go routine to insert a new hash into the map
// This must protect against simultaneous access from other goroutines.

func insertHashEvent(keyCounter int, sum [32]byte) {

        // start a duration to know how long this function takes to execute
        // (and so we won't get a pesk use of package time without selector
        // warning if we don't call an intrinsic function in the package)
        //
        startingTime := time.Now().UTC()

	fmt.Println("Now is : ", startingTime.Format(time.ANSIC))

	// get five minutes into future
	fiveMinutes := time.Minute * time.Duration(5)
	futureTime := startingTime.Add(fiveMinutes)
	var duration time.Duration = futureTime.Sub(startingTime)
	 
         fmt.Println("Five minutes from now will be : ", futureTime.Format(time.ANSIC))

        // Create the map on the first hash to be added.  Otherwise we panic.

        log.Printf("start keyCounter = %d\n", keyCounter)

	if (keyCounter == 0){
		log.Printf("insert: keyCounter = %d\n", keyCounter)
		mux.Lock()
                hashEventMap := make(map[int]*hashEventData)
		keyCounter++
		mux.Unlock()
		log.Printf("insert: keyCounter now = %d\n", keyCounter)
                i, ok := hashEventMap[1]
                if (ok) {
                        log.Printf("map initial query of entry 1 = %d", i)
		} else {

			log.Printf("current keyCounter = %d\n", keyCounter)

			mux.Lock()
			hashEventMap[keyCounter] = &hashEventData{duration,sum}
			mux.Unlock()
		}
	}
	return
}

func createHash(w http.ResponseWriter, r *http.Request) {

	var err error
	var size int
	var buf []byte
	var elapsedTime int64
	var sum [32]byte
	
	w.Write([]byte("Received a POST request\n"))

	log.Printf("URL.Path = %s\n", r.URL.Path)

	// start a duration to know how long this function takes to execute
	// (and so we won't get a pesk use of package time without selector
	// warning if we don't call an intrinsic function in the package)
	// 
//	var duration time.Duration = endingTime.Sub(startingTime)
//	elapsedTime = duration.Nanoseconds()/1e3

//	log.Printf("elapsedTime = %v usec\n", elapsedTime)
	startingTime := time.Now().UTC()

	w.Write([]byte("creating hashed password and inserting into map...\n"))

	// We need to read the request body in order to get the data 
	// passed in as a curl --data "password=xxxx" option.
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Create a temporary Bytes buffer to use to verify the password.
	// These Buffer calls consume the temporary buffer, which is freed
	// as each call is made.
        buf = []byte(reqBody)

	size = len(buf)

	// Check for minimum of "password=
	if (size < 9) {
		http.NotFound(w, r)
		return
	}
	// Create a sha256 sum and create a slice from the byte array 
	// at the password offset
	sum = sha256.Sum256(buf[10:size])
        fmt.Printf("%x\n", sum)

	go insertHashEvent(keyCounter, sum)

	// Stop the timer and use time.Duration to get the elapsed time.
	
	// This was really hard to figure out.  If you just call time.Now,
	// your build will fail with a "use of package time without selector"
	// failure.  This means you didn't use an intrinsic function in the
	// package (i.e. the ones cited in https://golang.org/pkg/time), 
	// the build error went away because I called time.Duration.

	// NOTE: the basic unit of time for duration is nanoseconds, 
	// and does not have a helper function like time.Millisecond
	// to convert to microseconds.

	endingTime := time.Now().UTC()

	var duration time.Duration = endingTime.Sub(startingTime)
	elapsedTime = duration.Nanoseconds()/1e3

	log.Printf("elapsedTime = %v usec\n", elapsedTime)

	w.Write([]byte("Received a POST request\n"))
}

// Add a show hash handler function.
// need to extract the key from the data.
func showHash(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL.Path = %s\n", r.URL.Path)
	w.Write([]byte("SHA256 for key...\n"))
}

/* 
 * Complete a showStats handler function by getting real numbers from
 * the map.
 *
 * go's json/Marshal package has a type reflector that translates types
 * from other languages into types support by goLang.  The reflector will
 * only accept type struct field names which begin with a capital letter.
 * If you use a lower case field name, no json encoding will be returned.
 * The workaround is to specify a field tag that indicates the name to be used.
 */
func showStats(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL.Path = %s\n", r.URL.Path)
	w.Write([]byte("JSON pair average, number\n")) 

	type Stat struct {
		Total     int `json:"total"`
		Average   int `json:"average"`
	}
	group := Stat{
		Total:     134,
		Average:   38,
	}
	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	// This will show the JSON string when running a curl command
	w.Write([]byte(b)) 
}

// Add a launchShutdown handler function.
func launchShutdown(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL.Path = %s\n", r.URL.Path)
	w.Write([]byte("Launching web server shutdown...\n"))
}
/*
 * Setup variables to be used by all goroutines needing
 * to access or update these data structures when processing
 * a request.  Updates to these variabes should use a mutex to 
 * protect them from race conditions with multiple go routines 
 * launched from multiple requests in flight.
 */

// Register the URL handler functions and corresponding 
// URL patterns with the NewServeMux
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", checkRootUrl)
	mux.HandleFunc("/hash", createHash)
	mux.HandleFunc("/hash/key", showHash)
	mux.HandleFunc("/stats", showStats)
	mux.HandleFunc("/shutdown", launchShutdown)


	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}