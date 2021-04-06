package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"runtime"
	"time"
)

var dummy_gui = fmt.Println
var button_color = color.RGBA{R: 100, G: 150, B: 210, A: 255}

/*****************************************************************************/
/*                         Create Menu                                       */
/*****************************************************************************/
func create_menu(a *taskApp) *fyne.MainMenu {
	importItem := fyne.NewMenuItem("Import", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				log.Println("Error: Import-1")
				a.show_error("Import", err)
				return
			}
			if reader == nil { // user cancelled
				log.Println("Import: User cancelled")
				return
			}

			log.Println("Reading tasks from import file: ", reader.URI().Path())
			a.readTasks(reader.URI())
			a.win.SetContent(a.makeUI())
		}, a.win)
	})

	exportItem := fyne.NewMenuItem("Export", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				log.Println("Error: Export-1")
				a.show_error("export tasks", err)
				return
			}
			if writer == nil { // user cancelled
				log.Println("Export: User cancelled")
				return
			}

			log.Println("Writing tasks to export file: ", writer.URI().Path())
			a.StoreListAsJSONFile(writer.URI(), false) // including deleted tasks
		}, a.win)
	})

	darkItem := fyne.NewMenuItem("Dark Theme", func() {
		a.app.Settings().SetTheme(theme.DarkTheme())
	})

	lightItem := fyne.NewMenuItem("Light Theme", func() {
		a.app.Settings().SetTheme(theme.LightTheme())
	})

	syncItem := fyne.NewMenuItem("Server-Sync", func() {
		a.settingsWindow = a.app.NewWindow("Server Settings")
		a.settingsWindow.SetContent(a.make_server_settings())
		a.settingsWindow.Resize(fyne.NewSize(480, 480))
		a.settingsWindow.Show()
	})

	aboutItem := fyne.NewMenuItem("About", func() {
		dialog.ShowInformation("About", about_text, a.win)
	})

	helpItem := fyne.NewMenuItem("Help", func() {
		a.helpWindow = a.app.NewWindow("Help")
		a.helpWindow.SetContent(a.make_help())
		if runtime.GOOS == "linux" {
			a.helpWindow.Resize(fyne.NewSize(480, 720))
		} else {
			a.helpWindow.SetFullScreen(true)
		}
		a.helpWindow.Show()
	})

	return fyne.NewMainMenu(
		fyne.NewMenu("File", importItem, exportItem),
		fyne.NewMenu("Settings", lightItem, darkItem, fyne.NewMenuItemSeparator(), syncItem),
		fyne.NewMenu("Help", aboutItem, helpItem),
	)
}

/*****************************************************************************/
/*                         Create Task Tab                                   */
/*****************************************************************************/
func create_task_tab(a *taskApp) *container.Scroll {

	/* -------------  Line with icons and show button ---------- */
	bar1 := widget.NewToolbar(
		// add new task
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			//due, _ := mydate_to_internal("31.12.30") // cannot use year 99, is regarded as 1999 (last century)
			due := time.Now().AddDate(0, 3, 0).Format("060102") // add 3 month
			t := &task{
				ID:          time.Now().UnixNano(),
				Name:        "New Task",
				Description: "",
				Duedate:     due,
				Ahead:       30,
				Category:    "",
				Owner:       "",
				Prio:        2,
			}
			t.State = state_future
			t.Done = false

			a.current = t
			a.tabbar.SelectTab(a.w_tab_details)
		}),

		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			// delete selected task
			if a.current == nil {
				a.show_error("No task selected", nil)
				return
			}
			s := fmt.Sprintf("Delete Task %s?", a.current.Name)
			dialog.ShowConfirm("Please confirm", s, func(v bool) {
				if v {
					a.current.Deleted = true
					a.current.ModTime = time.Now().Unix()
					a.f_stored = false
					a.show_tasks()
				}
			}, a.win)
		}),

		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			// copy content of selected task to a new task
			if a.current == nil {
				a.show_error("No task selected", nil)
				return
			}
			t := &task{
				ID:          time.Now().UnixNano(),
				Name:        fmt.Sprintf("Copy - %s", a.current.Name),
				Description: a.current.Description,
				Duedate:     internal_add(a.current.Duedate, 1, 0, 0),
				Ahead:       a.current.Ahead,
				Category:    a.current.Category,
				Owner:       a.current.Owner,
				Prio:        a.current.Prio,
			}
			t.State = a.current.State
			t.Done = a.current.Done

			a.current = t
			a.tabbar.SelectTab(a.w_tab_details)
		}),

		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			// edit selected task
			if a.current == nil {
				a.show_error("No task selected", nil)
				return
			}
			a.tabbar.SelectTab(a.w_tab_details)
		}),
	)

	tools1 := container.NewHBox(
		bar1,
	)

	a.w_b_taskshow = widget.NewButton("Show all", nil)
	a.w_b_taskshow.OnTapped = func() {
		if a.w_b_taskshow.Text == "Show all" {
			a.f_show_filtered = false
			a.w_listlabel.Text = "All Tasks"
			a.w_b_taskshow.SetText("Apply Filter")
		} else {
			a.f_show_filtered = true
			a.w_listlabel.Text = "Filtered Tasks"
			a.w_b_taskshow.SetText("Show all")
		}
		a.show_tasks()
	}

	c_show_all_filtered := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		a.w_b_taskshow,
	)

	bar_tasks := container.New(layout.NewBorderLayout(nil, nil, tools1, c_show_all_filtered), tools1, c_show_all_filtered)

	/* -------------  Task List ---------- */

	a.w_listlabel = widget.NewLabel("Filtered Tasks")

	a.w_tasklist = widget.NewList(
		func() int {
			return len(a.visible)
		},
		func() fyne.CanvasObject {
			return container.New(layout.NewMaxLayout(),
				canvas.NewRectangle(color.RGBA{G: 155, A: 100}),
				widget.NewLabel("Item x"))
		},
		func(i int, c fyne.CanvasObject) {
			task := a.visible[i]
			state, col := get_state_color(task)
			task.State = state

			due := get_formatted_duedate(task)
			box := c.(*fyne.Container)
			rect := box.Objects[0].(*canvas.Rectangle)
			rect.FillColor = col
			label := box.Objects[1].(*widget.Label)
			s := fmt.Sprintf("%s %s", due, task.Name)
			label.SetText(s)
		},
	)
	a.w_tasklist.OnSelected = func(id int) {
		a.current = a.visible[id]
	}

	/* -------------  Buttons at the end ---------- */

	button_save := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		widget.NewButton("Save", func() {
			dialog.ShowInformation("Saving: ", "Saving ...", a.win)

			a.StoreListAsJSONFile(a.file_uri, false) // including deleted tasks
			a.f_stored = true
		}),
	)
	button_quit := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		widget.NewButton("Quit", func() {
			if a.f_stored == false {
				dialog.ShowConfirm("Please confirm", "Exit without saving changes?", func(v bool) {
					if v {
						os.Exit(0)
					}
				}, a.win)
			} else {
				os.Exit(0)
			}
		}),
	)

	/* -------------  task container ---------- */
	inner_top := container.NewVBox(
		widget.NewLabel(""),
		bar_tasks,
		widget.NewLabel(""),
		container.New(layout.NewCenterLayout(), a.w_listlabel),
	)

	inner_bottum := button_save
	if runtime.GOOS == "linux" {
		inner_bottum = container.New(layout.NewBorderLayout(nil, nil, button_save, button_quit), button_save, button_quit)
	}

	return container.NewScroll(
		container.New(layout.NewBorderLayout(inner_top, inner_bottum, nil, nil),
			inner_top, inner_bottum, a.w_tasklist),
	)
}

/*****************************************************************************/
/*                         Create Details Tab                                */
/*****************************************************************************/
func create_details_tab(a *taskApp) *container.Scroll {
	bar_details := widget.NewToolbar(
		// save task
		widget.NewToolbarAction(theme.ConfirmIcon(), func() {
			a.f_stored = false
			a.SaveDetails()
		}),
		// show previous task
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			a.SetNextTask(-1)
			a.DisplayCurrentTask()
			a.tabbar.SelectTab(a.w_tab_details)
		}),
		// show next task
		widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
			a.SetNextTask(1)
			a.DisplayCurrentTask()
			a.tabbar.SelectTab(a.w_tab_details)
		}),
	)

	a.w_ID = widget.NewLabel("")
	a.w_modtime = widget.NewLabel("")
	a.w_name = widget.NewEntry()
	a.w_description = widget.NewEntry()
	a.w_duedate = widget.NewEntry() /* Details-Duedate is shown in mytime (e.g. 16.02.21) */
	a.w_ahead = widget.NewEntry()
	a.w_category = widget.NewEntry()
	a.w_owner = widget.NewEntry()
	a.w_prio = widget.NewEntry()

	details := widget.NewForm(
		widget.NewFormItem("Name", a.w_name),
		widget.NewFormItem("Description", a.w_description),
		widget.NewFormItem("Duedate", a.w_duedate),
		widget.NewFormItem("Ahead", a.w_ahead),
		widget.NewFormItem("Category", a.w_category),
		widget.NewFormItem("Owner", a.w_owner),
		widget.NewFormItem("Prio", a.w_prio),
	)

	a.w_details_done = widget.NewCheck("Done", func(v bool) {})

	box_details := container.NewScroll(container.NewVBox(
		widget.NewLabel(""),
		container.NewHBox(
			//widget.NewLabel ("save: "),
			bar_details,
		),
		container.New(layout.NewCenterLayout(), widget.NewLabel("Details")),
		container.NewHBox(widget.NewLabel("ID:       "), a.w_ID),
		container.NewHBox(widget.NewLabel("Mod-Time: "), a.w_modtime),
		details,
		a.w_details_done,
	))
	return box_details
}

/*****************************************************************************/
/*                         Create Filter Tab                                 */
/*****************************************************************************/
func create_filter_tab(a *taskApp) *container.Scroll {
	state_names := []string{"future", "soon", "now", "past", "done"}
	state_visible := []bool{false, true, true, true, false}

	a.filter.statuses = make([]bool, len(state_names))

	state_box := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for i := 0; i < len(state_names); i++ {
		k := i
		cb := widget.NewCheck(state_names[i], func(value bool) {
			a.filter.statuses[k] = value
		})
		cb.SetChecked(state_visible[i])
		state_box.Add(cb)
	}

	prio_names := []string{"1 high", "2 med", "3 low"}
	prio_visible := []bool{true, true, true}

	a.filter.prios = make([]bool, len(prio_names))

	prio_box := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for i := 0; i < len(prio_names); i++ {
		k := i
		cb := widget.NewCheck(prio_names[i], func(value bool) {
			a.filter.prios[k] = value
		})
		cb.SetChecked(prio_visible[i])
		prio_box.Add(cb)
	}

	w_sel_category := widget.NewSelect(a.category_list, func(s string) { a.filter.category = s })
	w_sel_category.SetSelected(a.category_list[0])

	w_sel_owner := widget.NewSelect(a.owner_list, func(s string) { a.filter.owner = s })
	w_sel_owner.SetSelected(a.owner_list[0])

	form := widget.NewForm(
		widget.NewFormItem("Status", state_box),
		widget.NewFormItem("Prio", prio_box),
		widget.NewFormItem("Category", w_sel_category),
		widget.NewFormItem("Owner", w_sel_owner),
	)

	label := container.New(layout.NewCenterLayout(), widget.NewLabel("Filter"))
	box := container.New(layout.NewBorderLayout(label, nil, nil, nil), label, form)

	return container.NewScroll(box)
}

/*****************************************************************************/
/*                         Create Sync Tab                                   */
/*****************************************************************************/
func create_sync_tab(a *taskApp) *fyne.Container {
	label := container.NewVBox(widget.NewLabel(""), widget.NewLabel("Sync-Log"))
	a.w_sync_log = widget.NewMultiLineEntry()
	b_start := widget.NewButton("Start", nil)
	b_cancel := widget.NewButton("Cancel", nil)

	b_start.OnTapped = func() {
		b_cancel.Enable()
		b_start.Disable()
		ret, err := a.do_sync()
		if err == nil {
			a.tasks = ret
			// now save to disk what was received from the server (=master) (drop deleted)
			a.StoreListAsJSONFile(a.file_uri, true)
			a.f_stored = true

			dia := dialog.NewInformation ("Sync", "Success", a.win)
            dia.SetOnClosed ( func() {
				a.show_tasks()
				a.tabbar.SelectTab(a.w_tab_tasks)
            })
            dia.Show()
		} else {
			dialog.ShowInformation("Sync", "Failed", a.win)
		}
		b_cancel.Disable()
		b_start.Enable()
	}
	b_cancel.OnTapped = func() {
		b_cancel.Disable()
		b_start.Enable()
	}

	c_start := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		b_start,
	)

	c_cancel := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		b_cancel,
	)

	b_box := container.New(layout.NewBorderLayout(nil, nil, c_start, c_cancel), c_start, c_cancel)

	container_log := container.NewVScroll(a.w_sync_log)

	return container.New(layout.NewBorderLayout(label, b_box, nil, nil), label, b_box, container_log)
}

/* =============================================================================================== */
func (a *taskApp) makeUI() fyne.CanvasObject {

	/* Create Menu */
	a.win.SetMainMenu(create_menu(a))

	/* task tab */
	box_tasks := create_task_tab(a)

	/* details tab */
	box_details := create_details_tab(a)

	/* filter tab */
	box_filter := create_filter_tab(a)

	/* sync tab */
	box_sync := create_sync_tab(a)

	s := "     Tasks" // on Android the menu icon covers the left tab item
	if runtime.GOOS == "linux" {
		s = "Tasks"
	}

	/* Tabs */
	a.w_tab_tasks = container.NewTabItem(s, box_tasks)
	a.w_tab_details = container.NewTabItem("Details", box_details)
	a.w_tab_filter = container.NewTabItem("Filter", box_filter)
	a.w_tab_sync = container.NewTabItem("Sync", box_sync)

	a.tabbar = container.NewAppTabs(a.w_tab_tasks, a.w_tab_details, a.w_tab_filter, a.w_tab_sync)
	a.tabbar.OnChanged = func(item *container.TabItem) {
		if item == a.w_tab_tasks {
			a.show_tasks()
		}
		if item == a.w_tab_details {
			a.DisplayCurrentTask()
		}
		if item == a.w_tab_sync {
			a.w_sync_log.SetText("")
		}
	}

	return a.tabbar
}

/* =============================================================================================== */
func (a *taskApp) make_server_settings() fyne.CanvasObject {

	w_server_ip := widget.NewEntry()
	w_server_ip.Text = a.server_ip
	w_server_port := widget.NewEntry()
	w_server_port.Text = a.server_port

	ip := widget.NewForm(
		widget.NewFormItem("IP4-Address", w_server_ip),
		widget.NewFormItem("Port", w_server_port),
	)

	apply := widget.NewButton("Apply", func() {
		a.server_ip = w_server_ip.Text
		a.server_port = w_server_port.Text
		a.storeSettings(a.settings_uri) // store in settings file for next start of app
		//a.settingsWindow.Close()                  // not needed, close via window-X button
	})

	c_button := fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		container.NewPadded(canvas.NewRectangle(button_color)),
		apply,
	)

	box_server := container.NewVBox(
		widget.NewLabel("Server Settings"),
		ip,
		c_button,
	)
	return box_server
}

/* =============================================================================================== */
func (a *taskApp) make_help() fyne.CanvasObject {

	e := widget.NewMultiLineEntry()
	e.Text = help_text
	e.Disable()
	e.Wrapping = fyne.TextWrapWord
	return e
}
