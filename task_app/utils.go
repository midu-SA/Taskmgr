package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"
)

const (

	// priority
	lowPriority  = 3
	midPriority  = 2
	highPriority = 1

	// status
	state_future = 0
	state_soon   = 1 // duedate minus ahead is less than 1 month away
	state_now    = 2
	state_past   = 3
	state_done   = 4
)

var dummy_utils = fmt.Println

/* --------------------------------------------------------------------------- */
func get_state_color(t *task) (state int, col color.NRGBA) {

	gray := color.NRGBA{B: 150, G: 150, R: 150, A: 250}
	red := color.NRGBA{R: 200, A: 250}
	green := color.NRGBA{G: 200, A: 250}
	orange := color.NRGBA{R: 200, G: 200, A: 250}

	duedate, err := time.Parse("060102", t.Duedate)
	if err != nil {
		log.Println("Parsing Error in duedate, Error: ", err)
		os.Exit(-1)
	}
	warndate := duedate.AddDate(0, 0, -1*t.Ahead)
	warndate_1m := warndate.AddDate(0, -1, 0)
	today := time.Now()

	if today.Before(warndate_1m) {
		state = state_future
	} else if today.Before(warndate) {
		state = state_soon
	} else if today.Before(duedate) {
		state = state_now
	} else if today.After(duedate) {
		state = state_past
	} else {
		state = state_now
	}

	if t.Done {
		state = state_done
	}

	if state == state_future || state == state_soon {
		col = green
	} else if state == state_now {
		col = orange
	} else if state == state_past {
		col = red
	} else if state == state_done {
		col = gray
	}

	return state, col
}

/* --------------------------------------------------------------------------- */
func get_formatted_duedate(t *task) string {
	proc_name := "get_formatted_duedate: "
	duedate, err := time.Parse("060102", t.Duedate)
	if err != nil {
		log.Println(proc_name, "Parsing Error in duedate, Error: ", err)
		os.Exit(-1) // todo: change
	}
	return duedate.Format("02.01.06")
}

/* ------------------------------------------------------------------------- */
/* convert   10.02.21   to 210210 */
func mydate_to_internal(t string) (string, error) {
	proc_name := "mydate_to_internal: "
	date, err := time.Parse("02.01.06", t)
	if err != nil {
		log.Println(proc_name, "Error: Parsing Error (", t, ")  Err: ", err)
		return "", err
	}
	return date.Format("060102"), nil
}

/* ------------------------------------------------------------------------- */
/* convert   210210  to  10.02.21 */
func internal_to_mydate(t string) string {
	proc_name := "internal_to_mydate: "
	date, err := time.Parse("060102", t)
	if err != nil {
		log.Println(proc_name, "Parsing Error (", t, ")  Err: ", err)
		os.Exit(-1) // todo change
	}
	return date.Format("02.01.06")
}

/* ------------------------------------------------------------------------- */
/* add years, months, days to a date formatted as dd.mm.yy */
func internal_add(t string, yy, mm, dd int) string {
	proc_name := "internal_add: "
	date, err := time.Parse("060102", t)
	if err != nil {
		log.Println(proc_name, "Parsing Error (", t, ")  Err: ", err)
		os.Exit(-1) // todo change
	}

	return date.AddDate(yy, mm, dd).Format("060102")
}

/* ------------------------------------------------------------------------- */
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

/* ------------------------------------------------------------------------- */
func make_category_list(tasks []*task) (li []string) {
	f_debug := false
	li = nil
	li = append(li, "all")
	if tasks != nil {
		for _, t := range tasks {
			if t == nil {
				break
			}
			if t.Deleted == true {
				continue
			}
			if t.Category == "" {
				continue
			}
			if !contains(li, t.Category) {
				li = append(li, t.Category)
			}
		}
	}
	if f_debug {
		log.Println("make_category_list: ", "Returning Category List: ", li)
	}
	f_debug = false
	return li
}

/* ------------------------------------------------------------------------- */
func make_owner_list(tasks []*task) (li []string) {
	f_debug := false
	li = nil
	li = append(li, "all")
	if tasks != nil {
		for _, t := range tasks {
			if t == nil {
				break
			}
			if t.Deleted == true {
				continue
			}
			if t.Owner == "" {
				continue
			}
			if !contains(li, t.Owner) {
				li = append(li, t.Owner)
			}
		}
	}
	if f_debug {
		log.Println("make_owner_list", "Returning Owner List: ", li)
	}
	f_debug = false
	return li
}
