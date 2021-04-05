# Taskmgr

This app can be used to manage tasks.
The app was tested on Linux and Android.

Tasks are displayed in a list with different colors, depending on when they are due.


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
(Usable only if a sync-server is running, see below)
Use the "Start" button to sync the tasks with an external server: all tasks are sent to the server (updated in the server) and a new task list is received. This new list is automatically stored on the internal file and then displayed.


----------- Menus -----------

** File-Menu **
Export: export your tasks to a file (e.g. for a backup)
Import: import your tasks from e.g. a backup file

** Settings-Menu **
Choose light or dark theme and set connections parameters (IP4 Address and Port) of your sync-server. 
The Apply Button saves the connection parameters in a local file.

** Help-Menu **
Show help text and app info.




#################################################################################

Build instructions:

1. Build the tasks app 

The app requires fyne at least in version 2.0.2  (https://fyne.io/)

1.1 Linux: 

- cd task_app
- go run *go

1.2 Android:

- copy *go tasks.png .../tasks
- cd .../tasks
- fyne package -os android -appID com.mdu.tasks -icon tasks.png 
- adb install tasks.apk


2. Build the sync server

my sync server runs on a raspberry pi. 
To build for the PI use e.g. the following commands:

- cd sync_server
- vi main.go (set the port your sync server is listening, here: 14000)
- GOOS=linux GOARCH=arm go build -o task_server  *.go
