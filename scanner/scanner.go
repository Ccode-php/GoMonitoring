package scanner

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"

	"GoMonitoring/models"
)

func pingHost(ip string) bool {

	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)

	return cmd.Run() == nil
}

func getHostname(ip string) string {

	names, err := net.LookupAddr(ip)

	if err != nil || len(names) == 0 {
		return ""
	}

	return strings.TrimSuffix(names[0], ".")
}

func getMac(ip string) string {

	_ = exec.Command("ping", "-c", "1", ip).Run()

	output, err := exec.Command(
		"ip",
		"neigh",
		"show",
		ip,
	).Output()

	if err != nil {
		return ""
	}

	fields := strings.Fields(string(output))

	for i, field := range fields {

		if field == "lladdr" && i+1 < len(fields) {

			mac := strings.ToUpper(fields[i+1])

			if mac == "" ||
				mac == "00:00:00:00:00:00" {

				return ""
			}

			return mac
		}
	}

	return ""
}

func ScanNetwork(network string) []models.Device {

	var (
		devices []models.Device
		wg      sync.WaitGroup
		mutex   sync.Mutex
	)

	baseIP := strings.TrimSuffix(network, "0/24")

	for i := 1; i <= 254; i++ {

		ip := fmt.Sprintf("%s%d", baseIP, i)

		wg.Add(1)

		go func(ip string) {

			defer wg.Done()

			if !pingHost(ip) {
				return
			}

			mac := getMac(ip)

			if mac == "" {
				return
			}

			device := models.Device{

				IP: ip,

				MAC: mac,
			}

			mutex.Lock()

			devices = append(
				devices,
				device,
			)

			mutex.Unlock()

			fmt.Printf(
				"FOUND %-15s %s\n",
				ip,
				mac,
			)

		}(ip)
	}

	wg.Wait()

	return devices
}
