package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var dummy_sync = fmt.Println

/* --------------------------------------------------------------------------------------------------- */
func (a *taskApp) do_sync() ([]*task, error) {

	proc_name := "do_sync: "
	s := ""
	a.w_sync_log.SetText(s)

	//url := "http://localhost:14000/synctasks"
	url := "http://" + a.server_ip + ":" + a.server_port + "/synctasks"

	body := a.tasks_to_json(false)

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Println("Failed to create a new request: ", err)
		s = s + "Failed to create new request: Err: " + err.Error() + "\n"
		a.w_sync_log.SetText(s)
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	s = s + "\nSending Post Request ...\n"
	a.w_sync_log.SetText(s)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(request)
	if err != nil {
		log.Println("Failed to send request: ", err)
		s = s + "Failed: Err: " + err.Error() + "\n"
		a.w_sync_log.SetText(s)
		return nil, err
	}
	defer resp.Body.Close()

	s = s + "\nReading Response...\n"
	a.w_sync_log.SetText(s)

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response: ", err)
		s = s + "Failed: Err: " + err.Error() + "\n"
		a.w_sync_log.SetText(s)
		return nil, err
	}

	statusCode := resp.StatusCode

	log.Println("Received: StatusCode: ", statusCode)
	//log.Println ("          Body:   ", string(resp_body))

	tasks := a.json_to_tasks(resp_body)

	msg := fmt.Sprintf("Received %d tasks", len(tasks))
	log.Println(proc_name, msg)

	s = s + msg + "\n"
	a.w_sync_log.SetText(s)

	return tasks, nil
}
