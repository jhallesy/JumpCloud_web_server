# JumpCloud_web_server
Programming assignment to implement a simple web server to create a golang http server with get/put requests

Go Web Server Design
Web Server Implementation

This is what I have so far.  Using the net/http package, I was able to implement function handlers  for all the expected requests, successfully implemented the SHA256 hash, and created a map indexed by an incrementing key counter, whose value held a struct that would have had the values needed for queries. 

The problem was the map I created in the hash handler did not persist because the handler returned.  You cannot take the address of a map in go, so it is impossible to pass it as a variable.
 
I figured out how to write a JSON reply for the /stats request, so it reports hard coded values as a JSON pair on stdout of the shell running a curl command.

The solution I arrived at after puzzling over this issue would have started a goroutine that created the map and had a context to send and receive information to allow other handlers to access the map it had created.  The goroutine would not return until a shutdown request was made.

The following is a description of the file go_web_server.go:

The server should respond to get and put requests on localhost:8080.  
As Get and Put keywords are not explicitly mentioned in the request, 
the functionality is done by inference from the URL pattern.

The request, issued by a user on the system hosting the go web server,
will use curl to send put and get commands to test the server’s responses.
This description of parsing the URL was the trickiest part of implementing 
the web server for this assignment.

Using the URL path to differentiate the methods can be done with an exact string match 
using the serveMux concept.  However, the exact match doesn’t work for the hash request
to obtain the key after a five second delay from creation using a counter value to fetch
the hash from a map.  The value is appended to the "hash" string of the request which uses 
a numerical value that can change.  By adding parsing code to the has request 
(“/hash/<some integer value>”, we can check to see if the string is a valid (hash, hashed-id)
	pair and call the showHash func (otherwise we reject the request to alert the user 
	they have submitted an invalid value). 

// Register the two new handler functions and corresponding URL
// patterns with thee serveMux
    mux := http.NewServeMux()
    mux.HandleFunc("/", home)
    mux.HandleFunc("/hash", createHash)
    mux.HandleFunc("/hash/number", showHash)
    mux.HandleFunc("/stats", showStats)
    mux.HandleFunc("/shutdown", launchShutdown)

Supported commands:
    1. curl -si -X POST —data “password=angryMonkey” http://localhost:8080/hash
       
The hash endpoint adds the password to a map as a key value pair; the KeyType value 
will be a unique counter value that increments with every POST request for a password. 
The ValueType struct referenced by the Key value will include a struct whose fields 
are a SHA512 encoded hashes, the time the hash was created, and completion time of the 
hash request from the time the request was called to when it finished, in microseconds. 
The additional timing metadata in the Value struct is needed by other requests such as
hash <identifier> or stats.

The Value struct is returned in the response to the request.  The amount of time (in microseconds) t
o complete each hash request before the reply is sent will be accumulated in a separate variable of type 
       
  2. curl -si -X GET http://localhost:8080/hash/<id>
       
When a GET request is made with the hash endpoint and id of a recently hashed password, 
the server will reply with “busy” until five seconds has passed, then it will reply
with the encoded hash string.  The spec calls for a SHA512 hash, but there is no support 
in the standard library for this so I implemented a 128 bit SHA256 checksum instead.
My notes:
       /*
        * Note: There is a SHA 512 byte sum function in the standard library:
        * func Sum512_256(data []byte) (sum256 [Size256]byte)
        * "Sum512_256 returns the Sum512/256 checksum of the data."
        *
        *      var data []byte
        *      err = Sum512_256(buf[10:size]) (data[0:255])
        */
3. curl -si -X POST  http://localhost:8080/stats

A GET request to /stats should return a JSON object with 2 key/value pairs. The 
“total” key should have a value for the count of POST requests to the /hash endpoint
made to the server so far. The “average” key should have a value for the average time 
it has taken to process all of those requests in microseconds.

For example: curl http://localhost:8080/stats should return something like: {“total”: 1, “average”: 123}

This was challenging, as the package time requires you to call one of the intrinsic function in the 
exported time package. If you just call time.Now, your build will fail with a "
use of package time without selector” warning.  
After some research, I learned if I called one of the top level functions defined in the time package, 
the error would go away.  I chose to use time.Duration.  
Figuring out how to format the time in microseconds was tricky too because there is no builtin 
conversion.  Duration time is in nanoseconds.

golang has a json encoder and I implemented a prototype to show how it could be used.
       
4. curl http://localhost:8080/shutdown
       
Provide support for a “graceful shutdown request”.
If a POST request is made to /shutdown, the server should reject new requests. 
The program should wait for any pending/in-flight work to finish before exiting. 

NOTE: this does not shutdown the OS!  The shutdown request would have communicated with 
the goroutine that manipulated the map structure I mentioned in the intro; I would have had 
shutdown set a flag telling the main function to stop calling ServeHttp handlers,
hen it would send a channel request to the long running go routine telling it to exit, 
and upon receiving the response the main loop would have exited.
	


