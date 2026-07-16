package models

type Port struct {
	Index int      `json:"index"`
	Name  string   `json:"name"`
	MACs  []string `json:"macs"`
}
