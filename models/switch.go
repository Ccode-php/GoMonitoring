package models

type Switch struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`

	Ports []Port `json:"ports"`
}
