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

func pingHost(ip string) bool {

	cmd := exec.Command(
		"ping",
		"-c", "1",
		"-W", "1",
		ip,
	)

	return cmd.Run() == nil
}

func getHostname(ip string) string {

	names, err := net.LookupAddr(ip)

	if err != nil || len(names) == 0 {

		return ""
	}

	return names[0]
}

func getMac(ip string) string {

	exec.Command(
		"ping",
		"-c",
		"1",
		ip,
	).Run()

	cmd := exec.Command(
		"ip",
		"neigh",
		"show",
		ip,
	)

	output, err := cmd.Output()

	if err != nil {
		return ""
	}

	line := string(output)

	fields := strings.Fields(line)

	for i, field := range fields {

		if field == "lladdr" {

			if len(fields) > i+1 {

				return fields[i+1]
			}
		}
	}

	return ""
}

func ScanNetwork(
	network string,
) []models.Device {

	var devices []models.Device

	var wg sync.WaitGroup

	var mutex sync.Mutex

	baseIP := strings.Replace(
		network,
		"0/24",
		"",
		1,
	)

	for i := 1; i <= 254; i++ {

		ip := fmt.Sprintf(
			"%s%d",
			baseIP,
			i,
		)

		wg.Add(1)

		go func(ip string) {

			defer wg.Done()

			fmt.Println(
				"Checking:",
				ip,
			)

			if pingHost(ip) {

				mac := getMac(ip)

				hostname := getHostname(ip)

				systemName,
					systemDescription,
					snmpEnabled :=
					snmp.GetSNMPInfo(ip)

				device := models.Device{

					IP: ip,

					MAC: mac,

					Hostname: hostname,

					SystemName: systemName,

					SystemDescription: systemDescription,

					SNMPEnabled: snmpEnabled,

					SNMPVersion: "2c",
				}

				mutex.Lock()

				devices = append(
					devices,
					device,
				)

				mutex.Unlock()

				fmt.Println(
					"FOUND:",
					ip,
				)
			}

		}(ip)
	}

	wg.Wait()

	return devices
}
