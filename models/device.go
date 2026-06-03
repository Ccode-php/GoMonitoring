package models

type Device struct {

	ID int `json:"id,omitempty"`

	IP string `json:"ip"`

	MAC string `json:"mac"`

	Hostname string `json:"hostname"`

	Vendor string `json:"vendor"`

	DeviceType string `json:"device_type"`

	SystemName string `json:"system_name"`

	SystemDescription string `json:"system_description"`

	SNMPEnabled bool `json:"snmp_enabled"`

	SNMPVersion string `json:"snmp_version"`
}