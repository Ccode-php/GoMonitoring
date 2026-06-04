package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"GoMonitoring/models"
)

const API_URL = "http://laravel:8000/api"

const TOKEN = "8xK92Lm@pQ!2026"

func SendDevices(
	devices []models.Device,
) error {

	jsonData, _ :=
		json.Marshal(devices)

	req, err := http.NewRequest(
		"POST",
		API_URL+"/scanner/report-batch",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"X-SCANNER-TOKEN",
		TOKEN,
	)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fmt.Println(
		"Laravel Status:",
		resp.Status,
	)

	return nil
}

type ScanTask struct {
	ID int `json:"id"`

	Network string `json:"network"`

	Status string `json:"status"`
}

func GetPendingTasks() (
	[]ScanTask,
	error,
) {

	resp, err :=
		http.Get(
			API_URL +
				"/scan-tasks/pending",
		)

	if err != nil {

		return nil, err
	}

	defer resp.Body.Close()

	var tasks []ScanTask

	err = json.NewDecoder(
		resp.Body,
	).Decode(&tasks)

	return tasks, err
}

func CompleteTask(
	id int,
) error {

	req, _ :=
		http.NewRequest(
			"POST",
			fmt.Sprintf(
				"%s/scan-tasks/%d/complete",
				API_URL,
				id,
			),
			nil,
		)

	client := &http.Client{}

	_, err :=
		client.Do(req)

	return err
}

func GetNetworks() (
	[]ScanTask,
	error,
) {

	resp, err :=
		http.Get(
			API_URL +
				"/scan-tasks/pending",
		)

	if err != nil {

		return nil, err
	}

	defer resp.Body.Close()

	var tasks []ScanTask

	err =
		json.NewDecoder(
			resp.Body,
		).Decode(
			&tasks,
		)

	return tasks, err
}
