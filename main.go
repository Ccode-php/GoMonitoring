package main

import (
	"fmt"
	"time"

	"GoMonitoring/api"
	"GoMonitoring/scanner"
)

func main() {

	for {

		networks, err :=
			api.GetNetworks()

		if err != nil {

			fmt.Println(
				"API ERROR:",
				err,
			)

			time.Sleep(
				time.Minute,
			)

			continue
		}

		for _, network := range networks {

			fmt.Println(
				"SCAN:",
				network.Network,
			)

			devices :=
				scanner.ScanNetwork(
					network.Network,
				)

			fmt.Println(
				"FOUND:",
				len(devices),
			)

			err :=
				api.SendDevices(
					devices,
				)

			if err != nil {

				fmt.Println(
					"SEND ERROR:",
					err,
				)
			}
		}

		fmt.Println(
			"WAIT 60 SECONDS",
		)

		time.Sleep(
			time.Minute,
		)
	}
}
