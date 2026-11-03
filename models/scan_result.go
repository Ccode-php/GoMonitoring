package models

type ScanResult struct {
	TaskID   int      `json:"task_id"`
	Devices  []Device `json:"devices"`
	Switches []Switch `json:"switches"`
}
