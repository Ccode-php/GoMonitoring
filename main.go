package main

import (
	"fmt"
	"time"

	"GoMonitoring/api"
	"GoMonitoring/scanner"
)

func main() {

	fmt.Println("Go Monitoring Scanner Started")

	for {

		networks, err := api.GetNetworks()

		if err != nil {

			fmt.Println("API ERROR:", err)

			time.Sleep(time.Minute)

			continue
		}

		if len(networks) == 0 {

			fmt.Println("No networks to scan")

			time.Sleep(time.Minute)

			continue
		}

		for _, network := range networks {

			fmt.Println("--------------------------------")
			fmt.Println("Scanning:", network.Network)

			devices := scanner.ScanNetwork(network.Network)

			fmt.Printf("Found %d device(s)\n", len(devices))

			if len(devices) == 0 {
				continue
			}

			if err := api.SendDevices(devices); err != nil {

				fmt.Println("SEND ERROR:", err)

				continue
			}

			fmt.Println("Devices sent successfully")
		}

		fmt.Println("--------------------------------")
		fmt.Println("Waiting 60 seconds...")
		fmt.Println()

		time.Sleep(time.Minute)
	}
}
