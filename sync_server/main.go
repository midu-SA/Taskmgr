package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var port string = "14000"
var dummy_main = fmt.Println

type task struct {
	ID          int64
	Name        string
	Description string
	Duedate     string // Format: 210830
	Ahead       int    // days to warn before due date
	Category    string
	Owner       string
	Prio        int
	State       int // internal (not stored in JSON File)
	Done        bool
	ModTime     int64 // last modification time as Unix time (nano-secs since 1970)
	Deleted     bool  // internal (if task was deleted)
}

var tasks  []*task
var ret_tasks []*task

var fn     string                  // servers filename to tasks
var f_busy bool

/* ---------------------------------------------------------------------------------------------- */
func main() {

	finish := make(chan bool)

	fn = "/tmp/tasks.json"
	tasks = ParseTaskFile(fn)
	if tasks == nil {
		fmt.Println("Warning: main: no tasks read")
	}

	go func() {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/synctasks", sync_tasks)
		http.ListenAndServe(":"+port, httpMux)
	}()

	<-finish
}

/* ---------------------------------------------------------------------------------------------- */
func sync_tasks(w http.ResponseWriter, r *http.Request) {
    f_debug := true
    if f_busy {
        log.Println("Server busy, Returning http-error 400")
		http.Error(w, "Server busy", 409)
		return
	}
     
    // Set Lock on server
    f_busy = true
	
    if r.Body == nil {
		log.Println("No request body received, Returning http-error 400")
		http.Error(w, "No request body received", 400)
        f_busy = false
		return
	}
	defer r.Body.Close()

	data, _ := ioutil.ReadAll(r.Body)
	//log.Println("- Header: ", r.Header)
	//log.Println("- Body: ", string(data))

	tasks_from_app := json_to_tasks(data)
    if f_debug {log.Println ("Received ", len(tasks_from_app), " tasks")}

	// loop through tasks and append to or delete or exchange in servers task list
	cnt_added := 0
	cnt_updated := 0
	for _, t := range tasks_from_app {
		ind := find_ID(t.ID)

		// append if not yet on server
		if ind == -1 {
			tasks = append(tasks, t)
            cnt_added++
			continue
		}

		// exchange if modification date is newer than on server
		if t.ModTime > tasks[ind].ModTime {
			tasks[ind] = t
			cnt_updated++
		}
	}

	if f_debug {
        log.Println ("sync: added: ", cnt_added, ", modified: ", cnt_updated)
    }
    
	// write new list to file
	if len(tasks) == 0 {
        log.Println ("Fatal Error: task list is empty")
        f_busy = false
        return
    }
	StoreListAsJSONFile (fn, tasks)
    
    // return non-deleted tasks
    ret_tasks = nil
    for _, t := range tasks {
        if t.Deleted == false {
            ret_tasks = append (ret_tasks, t)
        }
    }
    ret_data := tasks_to_json(ret_tasks)
	w.Header().Set("content-type", "application/json")
	//log.Println("Returning: ", string(ret_data))
	w.Write(ret_data)
    
    if f_debug {
        log.Println ("sync: returned to client: ", len(ret_tasks), " tasks")
    }
    f_busy = false
}

/* ---------------------------------------------------------------------------------------------- */
func find_ID(id int64) (ret int) {
    f_debug := false
    if f_debug {log.Println ("findID: Starting ...   id = ", id)}
	ret = -1
	for i, t := range tasks {
        if t== nil {continue}
		if t.ID == id {
			ret = i
			break
		}
	}
	if f_debug {log.Println ("findID: returning index: ", ret)}
	return ret
}
