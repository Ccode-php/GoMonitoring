package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"

	"GoMonitoring/models"
)

// Ping
func ping(ip string) bool {

	cmd := exec.Command(
		"ping",
		"-c",
		"1",
		"-W",
		"1",
		ip,
	)

	return cmd.Run() == nil
}

// Hostname
func getHostname(ip string) string {

	names, err := net.LookupAddr(ip)

	if err != nil || len(names) == 0 {

		return ""

	}

	return strings.TrimSuffix(
		names[0],
		".",
	)
}

// MAC Address
func getMAC(ip string) string {

	_ = exec.Command(
		"ping",
		"-c",
		"1",
		ip,
	).Run()

	out, err := exec.Command(
		"ip",
		"neigh",
		"show",
		ip,
	).Output()

	if err != nil {

		return ""

	}

	fields := strings.Fields(string(out))

	for i := 0; i < len(fields); i++ {

		if fields[i] == "lladdr" {

			if i+1 < len(fields) {

				return strings.ToUpper(
					fields[i+1],
				)

			}

		}

	}

	return ""
}

// Bitta hostni scan qilish
func scanHost(ip string) *models.Device {

	if !ping(ip) {

		return nil

	}

	device := &models.Device{

		IP: ip,

		MAC: getMAC(ip),

		Hostname: getHostname(ip),

		IsSwitch: false,
	}

	if device.MAC == "" {

		return nil

	}

	return device
}

// Network Scan
func Scan(network string) models.ScanResult {

	result := models.ScanResult{
		Devices:  make([]models.Device, 0),
		Switches: make([]models.Switch, 0),
	}

	var wg sync.WaitGroup

	var mutex sync.Mutex

	base := strings.TrimSuffix(
		network,
		"0/24",
	)

	for i := 1; i <= 254; i++ {

		ip := fmt.Sprintf(
			"%s%d",
			base,
			i,
		)

		wg.Add(1)

		go func(ip string) {

			defer wg.Done()

			device := scanHost(ip)

			if device == nil {

				return

			}

			// SNMP orqali switchni tekshirish
			sw, ok := GetSwitchInfo(ip)

			if ok {

				device.IsSwitch = true

				mutex.Lock()

				result.Switches = append(
					result.Switches,
					sw,
				)

				mutex.Unlock()
			}

			mutex.Lock()

			result.Devices = append(
				result.Devices,
				*device,
			)

			mutex.Unlock()

			fmt.Printf(
				"FOUND %-15s %s\n",
				device.IP,
				device.MAC,
			)

		}(ip)
	}

	wg.Wait()

	return result
}
