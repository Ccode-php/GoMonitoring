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

		tasks, err := api.GetNetworks()

		if err != nil {

			fmt.Println("API ERROR:", err)

			time.Sleep(time.Minute)

			continue
		}

		if len(tasks) == 0 {

			fmt.Println("No networks to scan")

			time.Sleep(time.Minute)

			continue
		}

		for _, task := range tasks {

			fmt.Println("--------------------------------")
			fmt.Println("Scanning:", task.Network)

			result := scanner.Scan(task.Network)

			fmt.Printf(
				"Devices: %d | Switches: %d\n",
				len(result.Devices),
				len(result.Switches),
			)

			if err := api.Send(result); err != nil {

				fmt.Println("SEND ERROR:", err)

				continue
			}

			fmt.Println("Scan completed")
		}

		fmt.Println("--------------------------------")
		fmt.Println("Waiting 60 seconds...")
		fmt.Println()

		time.Sleep(time.Minute)
	}
}
