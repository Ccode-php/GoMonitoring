package main

import (
	"fmt"
	"log"
	"time"

	"GoMonitoring/api"
	"GoMonitoring/scanner"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: .env file not found")
	}

	fmt.Println("Go Monitoring Scanner Started")

	for {

		config, err := api.GetConfig()

		if err != nil {

			fmt.Println(
				"CONFIG ERROR:",
				err,
			)

			time.Sleep(
				30 * time.Second,
			)

			continue
		}

		fmt.Printf(
			"Scanner config: interval=%ds SNMP=%s timeout=%ds retries=%d\n",
			config.ScanInterval,
			config.SNMPVersion,
			config.SNMPTimeout,
			config.SNMPRetries,
		)

		scanner.SetConfig(config)

		tasks, err := api.GetNetworks()

		if err != nil {

			fmt.Println(
				"API ERROR:",
				err,
			)

			time.Sleep(
				time.Duration(
					config.ScanInterval,
				) * time.Second,
			)

			continue
		}

		for _, task := range tasks {

			if !task.Enabled {
				continue
			}

			fmt.Println(
				"Scanning:",
				task.Network,
			)

			result := scanner.Scan(
				task.Network,
			)

			if err := api.Send(result); err != nil {

				fmt.Println(
					"SEND ERROR:",
					err,
				)

			}
		}

		time.Sleep(
			time.Duration(
				config.ScanInterval,
			) * time.Second,
		)
	}
}
