package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"log"
)

var dummy_main = fmt.Println

type S_FILTER struct {
	statuses []bool /* if true, status of list is selected: order = "future", "now", "past", "done" */
	prios    []bool /* if true, prio of list is selected: order = "unused", "high", "medium", "low" */
	category string /* possible to selected either a single category or all */
	owner    string /* possible to selected either a single owner or all */
}

type taskApp struct {
	app             fyne.App
	win             fyne.Window
	file_uri        fyne.URI /* JSON file with tasks */
	settings_uri    fyne.URI /* file with settings */
	tasks           []*task  /* all tasks */
	visible         []*task  /* displayed tasks according to filter criteria */
	f_show_filtered bool

	current *task /* a pointer to the selected or a new task */

	category_list []string
	owner_list    []string
	filter        S_FILTER

	w_tasklist                                                             *widget.List
	w_listlabel                                                            *widget.Label
	w_duedate, w_name, w_description, w_ahead, w_category, w_prio, w_owner *widget.Entry
	w_ID, w_modtime                                                        *widget.Label
	w_sel_category, w_sel_owner                                            *widget.Select
	w_sync_log                                                             *widget.Entry
	w_tab_tasks, w_tab_filter, w_tab_details, w_tab_sync                   *container.TabItem
	w_b_taskshow                                                           *widget.Button
	w_details_done                                                         *widget.Check

	tabbar   *container.AppTabs
	f_stored bool // false if changes were made, true after storing

	helpWindow     fyne.Window
	settingsWindow fyne.Window
	server_ip      string
	server_port    string
}

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

type settings struct {
	IP   string
	Port string
}

var about_text = `
Author: Michael Dupre
Version: 0.9

Required: fyne 2.0.2 or higher
`

var help_text = `
This app can be used to manage tasks.

Each task has parameters (which can be displayed in the Details tab), at least a name, a due date and an "ahead" number, which defaults to 30 days.
Different colors are used to indicate the status of a task.
- Green:  duedate is more than 
          "ahead" days away 
          (states future and soon)
- Yellow: duedate is less than 
          "ahead" days away (state now)
- Red:    duedate has passed already
          (state past)
- Grey:   task is marked as "done"

(Deleted tasks are not displayed, but kept in memory and synchronised with the server)


----------- Menus -----------

** File-Menu **
Export: export your tasks to a file (e.g. for a backup)
Import: import your tasks from e.g. a backup file

** Settings-Menu **
Choose light or dark theme and set connections parameters (IP4 Address and Port) of your sync-server. 
The Apply Button saves the connection parameters in a local file.

** Help-Menu **
Show help text and app info.
  

----------- Tabs ------------
  
** Tasks-Tab **
Tasks are displayed in a scrolled list. The button "Show all" resp. "Apply Filter" is used to show tasks, either all, or only tasks that match the filter criteria.
Icons can be used to add a new task (+), delete the currently selected task (-), copy and edit(modify) the currently selected task. 
The button "Save" is used to store all tasks (in an app-internal file). Note that without saving, all changes (add, delete, ...) are lost when the app is closed.
  
** Details-Tab **
Display and optionally change the details of the selected (or new) task. Use the icons to accept the changes, or display details of the previous/next task of the currently displayed list.
The "Done" checkbox can be used to mark/unmark a task as done.
Mandatory fields are pre-filled when the "add task" icon is used.
New owners and new categories can used, but these are only visible in the filter tab after restart of the app.

** Filter-Tab **
Set filter criteria, which apply if the "Apply Filter" button of the tasks-tab is used.
Status "soon"/"future": duedate minus adhead is less/more than 1 month away.
Filter criteria are: state, priority, category, owner.

** Sync-Tab **
(Usable only if a sync-server is running)
Use the "Start" button to sync the tasks with an external server: all tasks are sent to the server (updated in the server) and a new task list is received. This new list is automatically stored on the internal file and then displayed.

`

/* --------------------------------------------------------------------------------------------------- */
func main() {
	f_debug := true
	if f_debug {
		log.Println("main tasks: Starting ...")
	}
	app := app.NewWithID("Tasks")
	w := app.NewWindow("Task List")

	// Linux: RootURI returns /home/opc/.config/fyne/Tasks/
	uri1, err1 := storage.Child(app.Storage().RootURI(), "tasks.json")
	uri2, err2 := storage.Child(app.Storage().RootURI(), "settings.json")

	if err1 != nil {
		log.Println("tasks_main: error-1: ", err1)
	} else {
		log.Println("tasks_main: URI-1-String: ", uri1.String())
	}

	if err2 != nil {
		log.Println("tasks_main: error-2: ", err2)
	} else {
		log.Println("tasks_main: URI-2-String: ", uri2.String())
	}

	taskapp := &taskApp{app: app,
		win:             w,
		file_uri:        uri1,
		settings_uri:    uri2,
		f_stored:        true,
		f_show_filtered: true,

		server_ip:   "192.168.178.25",
		server_port: "12000",
	}

	taskapp.readTasks(uri1)
	taskapp.readSettings(uri2)

	taskapp.visible = taskapp.tasks
	w.SetContent(taskapp.makeUI())

	//w.SetFullScreen (true)
	taskapp.show_tasks()
	taskapp.tabbar.SelectTab(taskapp.w_tab_tasks)
	w.Resize(fyne.NewSize(400, 700))
	w.ShowAndRun()

}
