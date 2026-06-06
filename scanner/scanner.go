package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"

	"GoMonitoring/models"
	"GoMonitoring/snmp"
)

// -----------------------------
// PING
// -----------------------------
func pingHost(ip string) bool {
	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)
	return cmd.Run() == nil
}

// -----------------------------
// HOSTNAME
// -----------------------------
func getHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

// -----------------------------
// MAC ADDRESS (SAFE VERSION)
// -----------------------------
func getMac(ip string) string {

	// ARP table fill qilish uchun ping
	_ = exec.Command("ping", "-c", "1", ip).Run()

	cmd := exec.Command("ip", "neigh", "show", ip)
	output, err := cmd.Output()

	if err != nil {
		return ""
	}

	line := string(output)
	if line == "" {
		return ""
	}

	fields := strings.Fields(line)

	for i, f := range fields {
		if f == "lladdr" && i+1 < len(fields) {
			mac := fields[i+1]

			// basic validation
			if mac == "" || mac == "00:00:00:00:00:00" {
				return ""
			}

			return mac
		}
	}

	return ""
}

// -----------------------------
// SCAN NETWORK
// -----------------------------
func ScanNetwork(network string) []models.Device {

	var devices []models.Device
	var wg sync.WaitGroup
	var mutex sync.Mutex

	baseIP := strings.Replace(network, "0/24", "", 1)

	for i := 1; i <= 254; i++ {

		ip := fmt.Sprintf("%s%d", baseIP, i)

		wg.Add(1)

		go func(ip string) {
			defer wg.Done()

			if !pingHost(ip) {
				return
			}

			mac := getMac(ip)

			// 🔥 CRITICAL FIX
			if mac == "" {
				mac = "UNKNOWN"
			}

			hostname := getHostname(ip)

			systemName, systemDescription, snmpEnabled :=
				snmp.GetSNMPInfo(ip)

			device := models.Device{
				IP:                ip,
				MAC:               mac,
				Hostname:          hostname,
				SystemName:        systemName,
				SystemDescription: systemDescription,
				SNMPEnabled:       snmpEnabled,
				SNMPVersion:       "2c",
			}

			mutex.Lock()
			devices = append(devices, device)
			mutex.Unlock()

			fmt.Println("FOUND:", ip)

		}(ip)
	}

	wg.Wait()

	return devices
}
