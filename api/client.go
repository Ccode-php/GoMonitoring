package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"GoMonitoring/models"
)

func getAPIURL() string {
	return os.Getenv("API_URL")
}

func getToken() string {
	return os.Getenv("SCANNER_TOKEN")
}

type ScanTask struct {
	ID      int    `json:"id"`
	Network string `json:"network"`
	Enabled bool   `json:"enabled"`
}

type ScannerConfig struct {
	ScanInterval      int    `json:"scan_interval"`
	OfflineTimeout    int    `json:"offline_timeout"`
	NotificationSound bool   `json:"notification_sound"`
	AutoRefresh       bool   `json:"auto_refresh"`
	SNMPCommunity     string `json:"snmp_community"`
	SNMPVersion       string `json:"snmp_version"`
	SNMPTimeout       int    `json:"snmp_timeout"`
	SNMPRetries       int    `json:"snmp_retries"`
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
		getAPIURL()+"/scan-tasks/pending",
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(
		"X-SCANNER-TOKEN",
		getToken(),
	)

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

// Laravel'dan scanner konfiguratsiyasini oladi
func GetConfig() (ScannerConfig, error) {

	var config ScannerConfig

	req, err := http.NewRequest(
		http.MethodGet,
		getAPIURL()+"/scanner/config",
		nil,
	)

	if err != nil {
		return config, err
	}

	req.Header.Set(
		"X-SCANNER-TOKEN",
		getToken(),
	)

	resp, err := client().Do(req)

	if err != nil {
		return config, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return config, fmt.Errorf(
			"config API returned %s",
			resp.Status,
		)
	}

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

// Scan natijasini Laravelga yuboradi
func Send(result models.ScanResult) error {

	body, err := json.Marshal(result)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		getAPIURL()+"/scanner/report-batch",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(
		"X-SCANNER-TOKEN",
		getToken(),
	)

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
