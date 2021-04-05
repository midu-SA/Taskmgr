package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"io/ioutil"
	"log"
	"time"
)

type JSON_TASK struct {
	ID          int64  `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description,omitempty"`
	Duedate     string `json:"Duedate"`
	Ahead       int    `json:"Ahead"`
	Category    string `json:"Category,omitempty"`
	Owner       string `json:"Owner,omitempty"`
	Prio        int    `json:"Prio,omitempty"`
	Done        bool   `json:"Done,omitempty"`
	ModTime     int64  `json:"ModTime,omitempty"`
	Deleted     bool   `json:"Deleted"`
}

type JSON_SETTINGS struct {
	IP   string `json:"IP"`
	Port string `json:"Port"`
}

var dummy = fmt.Println
var maxno int = 2000

/* ========================================================================================== */
/* read JSON History File and return task list */
func (a *taskApp) ParseTaskFile(uri fyne.URI) []*task {
	proc_name := "ParseTaskFile: "
	f_debug := false

	if f_debug {
		log.Println(proc_name, "Starting ...  (URL: ", uri.String(), ")")
	}

	fn := uri.String()

	v, err := storage.CanRead(uri)
	if v == false {
		dialog.ShowInformation("Tasks-File", "Does not yet exist. Add new Tasks!", a.win)
		return nil
	}
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot read JSON file %s", fn), err)
		return nil
	}

	read, err := storage.Reader(uri)
	if err != nil {
		a.show_error("OpenFileFromURI", err)
		return nil
	}
	defer read.Close()

	data, err := ioutil.ReadAll(read)
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot read JSON file %s", fn), err)
		return nil
	}

	if f_debug {
		s := fmt.Sprintf("%s Read Bytes: <%s>", proc_name, string(data))
		log.Println(proc_name, s)
	}

	tasks := a.json_to_tasks(data)
	return tasks
}

/* ========================================================================================== */
func (a *taskApp) StoreListAsJSONFile(uri fyne.URI, f_drop_deleted bool) {
	proc_name := "StoreListAsJSONFile: "
	f_debug := false
	if f_debug {
		log.Println(proc_name, "Starting ...")
	}

	v, err := storage.CanWrite(uri)
	if v == false {
		dialog.ShowInformation("Tasks-File", "Cannot write!", a.win)
		return
	}

	if err != nil {
		a.show_error("storage.CanWrite", err)
		return
	}

	storer, err := storage.Writer(uri)
	if err != nil {
		a.show_error("Storage.Writer", err)
		return
	}
	defer storer.Close()

	data := a.tasks_to_json(f_drop_deleted)

	n, err := storer.Write(data)
	if f_debug {
		log.Println(proc_name, n, " Bytes written: ")
	}
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot write to JSON file"), err)
		return
	}

	if f_debug {
		s := fmt.Sprintf("%d Bytes written: <%s>", n, string(data))
		log.Println(proc_name, s)
	}
}

/* ========================================================================================== */
func (a *taskApp) json_to_tasks(data []byte) []*task {
	proc_name := "json_to_tasks: "
	f_debug := false

	var tasks []*task
	tasks = nil

	p := make([]JSON_TASK, maxno)
	err := json.Unmarshal(data, &p)
	if err != nil {
		a.show_error(fmt.Sprintf("Parsing error in JSON file"), err)
		dialog.ShowInformation("", "Starting with empty task list", a.win)
		return nil
	}

	num := 0

	for i, ta := range p {
		if i >= maxno {
			a.show_error("Max num of tasks reached (2000)", nil)
			break
		}

		if ta.Name == "" {
			log.Println("ParseFile: Should never be reached!  i: ", i)
			continue
		}

		due, err := mydate_to_internal(ta.Duedate) // convert duedate to internal format  yymmdd
		if err != nil {
			a.show_error(fmt.Sprintf("Wrong duedate in task %s ... skip task", ta.Name), nil)
			continue
		}

		if ta.Name == "" {
			break
		}

		t := &task{
			ID:          ta.ID,
			Name:        ta.Name,
			Description: ta.Description,
			Duedate:     due,
			Ahead:       ta.Ahead,
			Category:    ta.Category,
			Owner:       ta.Owner,
			Prio:        ta.Prio,
			Done:        ta.Done,
			ModTime:     ta.ModTime,
			Deleted:     ta.Deleted,
		}

		/* for migration only */
		if t.ID == 0 {
			t.ID = time.Now().UnixNano()
		}
		if t.ModTime == 0 {
			t.ModTime = time.Now().Unix()
		}

		t.State, _ = get_state_color(t) // requires duedate in internal format
		if ta.Prio == 0 {
			t.Prio = 2 // Default: 2 = medium
		}

		tasks = append(tasks, t)
		num++
	}

	if f_debug {
		for i := 0; i < num; i++ {
			log.Println(proc_name, "Name: ", tasks[i].Name, ", Duedate: ", tasks[i].Duedate)
		}
	}
	return tasks
}

/* ========================================================================================== */
func (a *taskApp) tasks_to_json(f_drop_deleted bool) []byte {
	proc_name := "tasks_to_json: "
	f_debug := false

	li := a.tasks
	num := len(li)
	if f_debug {
		log.Println(proc_name, "Num: ", num)
	}

	p := make([]JSON_TASK, num)
	k := 0 // counter for JSON struct
	for i := 0; i < num; i++ {
		if f_drop_deleted && li[i].Deleted {
			continue
		}
		if li[i].Name == "" {
			continue // should never be reached
		}

		p[k].ID = li[i].ID
		p[k].ModTime = li[i].ModTime
		p[k].Name = li[i].Name
		p[k].Description = li[i].Description
		p[k].Duedate = internal_to_mydate(li[i].Duedate)
		p[k].Ahead = li[i].Ahead
		p[k].Category = li[i].Category
		p[k].Owner = li[i].Owner
		p[k].Prio = li[i].Prio
		p[k].Done = li[i].Done
		p[k].Deleted = li[i].Deleted
		k++
	}

	b, err := json.MarshalIndent(p[0:k], "", "  ")

	if err != nil {
		a.show_error("Mashal-JSON", err)
		return nil
	}
	return b
}

/* ========================================================================================== */
/*                                Settings                                                    */
/* ========================================================================================== */

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) readSettings(uri fyne.URI) {
	proc_name := "readSettings: "

	f_debug := false
	var set settings

	v, err := storage.CanRead(uri)
	if v == false {
		dialog.ShowInformation("Settings-File", "Does not yet exist", a.win)
		return
	}
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot read Settings file"), err)
		return
	}

	read, err := storage.Reader(uri)
	if err != nil {
		a.show_error("reader settings", err)
		return
	}
	defer read.Close()

	data, err := ioutil.ReadAll(read)
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot read Settings file"), err)
		return
	}

	if f_debug {
		s := fmt.Sprintf("Read Bytes: <%s>", string(data))
		log.Println(proc_name, s)
	}

	err = json.Unmarshal(data, &set)
	if err != nil {
		a.show_error(fmt.Sprintf("Parsing error in settings file"), err)
		return
	}

	a.server_ip = set.IP
	a.server_port = set.Port
	return
}

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) storeSettings(uri fyne.URI) {
	proc_name := "storeSettings: "

	f_debug := false
	var set JSON_SETTINGS

	v, err := storage.CanWrite(uri)
	if v == false {
		dialog.ShowInformation("Settings-File", "Cannot write!", a.win)
		return
	}

	if err != nil {
		a.show_error("storage.CanWrite", err)
		return
	}

	storer, err := storage.Writer(uri)
	if err != nil {
		a.show_error("Storage.Writer", err)
		return
	}
	defer storer.Close()

	set.IP = a.server_ip
	set.Port = a.server_port

	data, err := json.MarshalIndent(set, "", "  ")

	if err != nil {
		a.show_error("Mashal-JSON", err)
		return
	}

	n, err := storer.Write(data)
	if f_debug {
		log.Println(proc_name, n, " Bytes written: ")
	}
	if err != nil {
		a.show_error(fmt.Sprintf("Cannot write to Settings file"), err)
		return
	}

	if f_debug {
		s := fmt.Sprintf("%d Bytes written: <%s>", n, string(data))
		log.Println(proc_name, s)
	}

	return
}
