package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"GoMonitoring/models"
)

type ScanTask struct {
	ID      int    `json:"id"`
	Network string `json:"network"`
	Enabled bool   `json:"enabled"`
}

type ScannerConfig struct {
	ScanInterval  int    `json:"scan_interval"`
	SNMPCommunity string `json:"snmp_community"`
	SNMPVersion   string `json:"snmp_version"`
	SNMPTimeout   int    `json:"snmp_timeout"`
	SNMPRetries   int    `json:"snmp_retries"`
}

// Bitta umumiy client (ulanishlar qayta ishlatiladi)
var httpClient = &http.Client{Timeout: 60 * time.Second}

func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	base := strings.TrimRight(os.Getenv("API_URL"), "/")

	req, err := http.NewRequest(method, base+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-SCANNER-TOKEN", os.Getenv("SCANNER_TOKEN"))

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// do so'rovni yuboradi, 200 bo'lmasa xato qaytaradi, out != nil bo'lsa JSON'ni dekod qiladi.
func do(req *http.Request, out any) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s %s -> %s: %s",
			req.Method, req.URL.Path, resp.Status, strings.TrimSpace(string(msg)))
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// Laravel'dan scan qilinadigan tarmoqlarni oladi
func GetNetworks() ([]ScanTask, error) {
	req, err := newRequest(http.MethodGet, "/scan-tasks/pending", nil)
	if err != nil {
		return nil, err
	}

	var tasks []ScanTask
	if err := do(req, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// Laravel'dan scanner konfiguratsiyasini oladi
func GetConfig() (ScannerConfig, error) {
	var cfg ScannerConfig

	req, err := newRequest(http.MethodGet, "/scanner/config", nil)
	if err != nil {
		return cfg, err
	}

	err = do(req, &cfg)
	return cfg, err
}

// Scan natijasini Laravelga yuboradi
func Send(result models.ScanResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}

	req, err := newRequest(http.MethodPost, "/scanner/report-batch", bytes.NewReader(body))
	if err != nil {
		return err
	}

	return do(req, nil)
}
