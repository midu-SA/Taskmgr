package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"log"
	"strconv"
	"time"
)

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) readTasks(uri fyne.URI) {
	proc_name := "readTasks: "
	f_debug := true
	if f_debug {
		log.Println(proc_name, "URI: ", uri.String())
	}

	a.tasks = a.ParseTaskFile(uri)
	a.category_list = make_category_list(a.tasks)
	a.owner_list = make_owner_list(a.tasks)
	a.current = &task{}
	a.visible = nil
	return
}

/* --------------------------------------------------------------------------------------------------- */
/* Save change made in details tab (add/modify task) */
func (a *taskApp) SaveDetails() {
	prio, err := strconv.Atoi(a.w_prio.Text)
	if err != nil || prio < 0 || prio > 3 {
		a.show_error("", errors.New("Prio must be 1,2 or 3"))
		return
	}
	ahead, err := strconv.Atoi(a.w_ahead.Text)
	if err != nil || ahead < 0 || ahead > 10000 {
		a.show_error("", errors.New("ahead must be > 0 and < 10000"))
		return
	}

	myd, err := mydate_to_internal(a.w_duedate.Text)
	if err != nil {
		a.show_error("", errors.New("wrong date or wrong format (<dd.mm.yy>)"))
		return
	}

	a.current.Name = a.w_name.Text
	a.current.Description = a.w_description.Text
	a.current.Duedate = myd
	a.current.Ahead = ahead
	a.current.Category = a.w_category.Text
	a.current.Owner = a.w_owner.Text
	a.current.Prio = prio
	a.current.ModTime = time.Now().Unix()

	if a.w_details_done.Checked {
		a.current.Done = true
		a.current.State = state_done
	} else {
		a.current.Done = false
	}

	/* re-create lists with valid categories and owners, in case owner or category was changed */
	/* Note: this does not work: options of an existing Select widget cannot be changed */

	f_found := false
	for _, t := range a.tasks {
		if *a.current == *t {
			f_found = true
			break
		}
	}

	if f_found == false {
		a.tasks = append(a.tasks, a.current)
	}

	a.tabbar.SelectTab(a.w_tab_tasks)
	return
}

/* --------------------------------------------------------------------------------------------------- */
/* return task list to display */
func (a *taskApp) build_task_list() (li []*task) {
	f_debug := true
	a.current = nil
	for _, t := range a.tasks {
		if t == nil {
			continue
		}
		if t.Deleted == true {
			continue
		}
		if a.f_show_filtered == true && t.matchFilter(a.filter) {
			li = append(li, t)
		}
		if a.f_show_filtered == false {
			li = append(li, t)
		}
	}
	if f_debug {
		log.Println("show_tasks", "Visible Task Count: ", len(li))
	}
	f_debug = false
	return li
}

/* --------------------------------------------------------------------------------------------------- */
/* show all tasks according to filter criteria, select first entry and set current to first task in list */
func (a *taskApp) show_tasks() {
	a.visible = a.build_task_list()
	if a.visible == nil {
		return
	}
	a.w_tasklist.Refresh()
	a.w_tasklist.Select(0) // select first item and update Details
	a.current = a.visible[0]
}

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) DisplayCurrentTask() {
	if a.current == nil {
		return
	} // should never happen
	t := a.current

	dt := time.Unix(t.ModTime, 0)
	curtime := fmt.Sprintf("%s", dt.Format("02.01.2006 15:04:05"))

	a.w_ID.SetText(strconv.FormatInt(t.ID, 10))
	if t.ModTime == 0 {
		a.w_modtime.SetText("")
	} else {
		a.w_modtime.SetText(curtime)
	}
	a.w_name.SetText(t.Name)

	a.w_description.SetText(t.Description)
	a.w_duedate.SetText(internal_to_mydate(t.Duedate))
	a.w_ahead.SetText(fmt.Sprintf("%d", t.Ahead))
	a.w_category.SetText(t.Category)
	a.w_owner.SetText(t.Owner)
	a.w_prio.SetText(fmt.Sprintf("%d", t.Prio))
	if t.Done {
		a.w_details_done.SetChecked(true)
	} else {
		a.w_details_done.SetChecked(false)
	}
}

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) SetNextTask(incr int) {
	proc_name := "SetNextTask: "
	f_debug := true
	le := len(a.visible)
	if le == 0 {
		a.current = nil
		return
	}

	f_found := false
	found_index := -1
	for i, t := range a.visible {
		if *a.current == *t {
			f_found = true
			found_index = i
			break
		}
	}
	if f_found == false {
		a.current = a.visible[0]
		return
	}

	if f_debug {
		log.Println(proc_name, "Foundindex: ", found_index)
	}
	ind := found_index + incr
	if ind < 0 {
		a.current = a.visible[le-1]
	} else if ind >= le {
		a.current = a.visible[0]
	} else {
		a.current = a.visible[ind]
	}
	if f_debug {
		log.Println(proc_name, "Found Task: ", a.current.Name)
	}
	return
}

/* ----------------------------------------------------------------------------------------------- */
func (a *taskApp) show_error(s string, err error) {
	if err != nil {
		s = fmt.Sprintf("%s, Error: %s", s, err.Error())
	}
	log.Println("Error: ", s)
	dialog.ShowError(errors.New(s), a.win)
}

/* --------------------------------------------------------------------------------------------------- */
func (t *task) matchFilter(filter S_FILTER) (ret bool) {
	proc_name := "task.matchFilter: "
	f_debug := false

	ret = true
	if t == nil {
		f_debug = false
		return false
	}

	if f_debug {
		log.Println(proc_name, "Checking Task ", t.Name)
	}
	if filter.statuses[t.State] == false {
		ret = false
	}

	if filter.prios[t.Prio-1] == false {
		ret = false
	}

	if filter.category != "all" && filter.category != t.Category {
		ret = false
	}

	if filter.owner != "all" && filter.owner != t.Owner {
		ret = false
	}

	if f_debug && ret == true {
		log.Println(proc_name, "Returning true")
	}
	f_debug = false
	return ret
}
