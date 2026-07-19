package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"GoMonitoring/api"
	"GoMonitoring/scanner"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	fmt.Println("Go Monitoring Scanner Started")

	interval, err := strconv.Atoi(
		os.Getenv("SCAN_INTERVAL"),
	)

	if err != nil || interval <= 0 {
		interval = 60
	}

	for {

		tasks, err := api.GetNetworks()

		if err != nil {

			fmt.Println("API ERROR:", err)

			time.Sleep(
				time.Duration(interval) * time.Second,
			)

			continue
		}

		for _, task := range tasks {

			result := scanner.Scan(task.Network)

			if err := api.Send(result); err != nil {

				fmt.Println(err)

			}

		}

		time.Sleep(
			time.Duration(interval) * time.Second,
		)

	}
}
