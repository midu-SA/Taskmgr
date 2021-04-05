package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

var dummy = fmt.Println
var maxno int = 2000

/* ========================================================================================== */
/* read JSON History File and return task list */
func ParseTaskFile(fn string) []*task {
    f_debug := false
    if f_debug {
        log.Println ("Reading tasks file...")
    }
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("Cannot read file ", fn, " Error: ", err)
		return nil
	}

	tasks := json_to_tasks(data)
	return tasks
}

/* ========================================================================================== */
// write list to local task file
func StoreListAsJSONFile(fn string, tasklist []*task)  {
    f_debug := false
    if f_debug {
        log.Println ("Writing tasks file...")
    }
    
	data := tasks_to_json(tasklist)
	err := ioutil.WriteFile(fn, data, 0644)
	if err != nil {
		log.Println("Cannot write file ", fn, " Error: ", err)
		return
	}
	return
}


/* ========================================================================================== */
func json_to_tasks(data []byte) []*task {
	proc_name := "json_to_tasks: "
	f_debug := false

	var tasks []*task
	tasks = nil

	p := make([]JSON_TASK, maxno)
	err := json.Unmarshal(data, &p)
	if err != nil {
		fmt.Println("Parsing error in JSON file", err)
		return nil
	}

	num := 0

	for i, ta := range p {
		if i >= maxno {
			log.Println("json_to_tasks: MAXNO reached")
			break
		}

		t := &task{
			ID:          ta.ID,
			Name:        ta.Name,
			Description: ta.Description,
			Duedate:     ta.Duedate,
			Ahead:       ta.Ahead,
			Category:    ta.Category,
			Owner:       ta.Owner,
			Prio:        ta.Prio,
			Done:        ta.Done,
			ModTime:     ta.ModTime,
			Deleted:     ta.Deleted,
		}

		tasks = append(tasks, t)
		num++
	}

	if f_debug {
		for i := 0; i < num; i++ {
			log.Println(proc_name, "Name: ", tasks[i].Name, ", Duedate: ", tasks[i].Duedate)
		}
	}
	if f_debug {log.Println (proc_name, "Returning...")}
	return tasks
}

/* ========================================================================================== */
func tasks_to_json(tasklist []*task) []byte {
	proc_name := "tasks_to_json: "
	f_debug := false

	num := len(tasklist)
	if f_debug {
		log.Println(proc_name, "Num: ", num)
	}

	p := make([]JSON_TASK, num)
	k := 0 // counter for JSON struct
	for i := 0; i < num; i++ {
		if tasklist[i] == nil {
			continue
		}
        if tasklist[i].Deleted == true {
			continue
		}
		p[k].ID = tasklist[i].ID
		p[k].ModTime = tasklist[i].ModTime
		p[k].Name = tasklist[i].Name
		p[k].Description = tasklist[i].Description
		p[k].Duedate = tasklist[i].Duedate
		p[k].Ahead = tasklist[i].Ahead
		p[k].Category = tasklist[i].Category
		p[k].Owner = tasklist[i].Owner
		p[k].Prio = tasklist[i].Prio
		p[k].Done = tasklist[i].Done
		p[k].Deleted = tasklist[i].Deleted
		k++
	}

	b, err := json.MarshalIndent(p[0:k], "", "  ")

	if err != nil {
		log.Println("tasks_to_json: Mashal-JSON: Error: ", err)
		return nil
	}
	return b
}
