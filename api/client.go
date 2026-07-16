package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"GoMonitoring/models"
)

const (
	//APIURL = "http://10.43.60.110:8000/api"
	APIURL = "http://127.0.0.1:8000/api"
	Token  = "8xK92Lm@pQ!2026"
)

type ScanTask struct {
	ID      int    `json:"id"`
	Network string `json:"network"`
	Enabled bool   `json:"enabled"`
}

func client() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// Laraveldan scan qilinadigan tarmoqlarni oladi
func GetNetworks() ([]ScanTask, error) {

	req, err := http.NewRequest(
		http.MethodGet,
		APIURL+"/scan-tasks/pending",
		nil,
	)

	if err != nil {
		return nil, err
	}

	resp, err := client().Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var tasks []ScanTask

	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Scan natijasini Laravelga yuboradi
func Send(result models.ScanResult) error {

	body, err := json.Marshal(result)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		APIURL+"/scanner/report-batch",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SCANNER-TOKEN", Token)

	resp, err := client().Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return fmt.Errorf(
			"laravel returned %s",
			resp.Status,
		)

	}

	fmt.Println("Scan result sent.")

	return nil
}
