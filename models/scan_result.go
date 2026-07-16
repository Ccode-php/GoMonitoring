package models

type ScanResult struct {
	Devices  []Device `json:"devices"`
	Switches []Switch `json:"switches"`
}
