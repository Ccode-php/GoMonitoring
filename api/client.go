package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"GoMonitoring/models"
)

const API_URL = "http://10.43.60.110:8000/api"
const TOKEN = "8xK92Lm@pQ!2026"

type ScanTask struct {
	ID      int    `json:"id"`
	Network string `json:"network"`
	Status  string `json:"status"`
}

// Laravel'dan scan qilinadigan tarmoqlarni oladi
func GetNetworks() ([]ScanTask, error) {

	resp, err := http.Get(API_URL + "/scan-tasks/pending")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var tasks []ScanTask

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Topilgan qurilmalarni Laravel'ga yuboradi
func SendDevices(devices []models.Device) error {

	jsonData, err := json.Marshal(devices)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		API_URL+"/scanner/report-batch",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SCANNER-TOKEN", TOKEN)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fmt.Println("Laravel Status:", resp.Status)

	return nil
}
