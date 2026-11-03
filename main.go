package main

import (
	"log"
	"time"

	"GoMonitoring/api"
	"GoMonitoring/scanner"

	"github.com/joho/godotenv"
)

const (
	minInterval = 5 * time.Second
	errorWait   = 30 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env topilmadi, muhit o'zgaruvchilari ishlatiladi")
	}

	log.Println("Go Monitoring Scanner started")

	for {
		started := time.Now()
		interval := runOnce()

		// Interval scan boshlanishidan hisoblanadi (scan davomiyligi ham kiradi)
		wait := interval - time.Since(started)
		if wait < time.Second {
			wait = time.Second
		}
		time.Sleep(wait)
	}
}

// runOnce bitta to'liq aylanishni bajaradi va keyingi aylanish intervalini qaytaradi.
func runOnce() time.Duration {
	cfg, err := api.GetConfig()
	if err != nil {
		log.Println("CONFIG ERROR:", err)
		return errorWait
	}

	interval := time.Duration(cfg.ScanInterval) * time.Second
	if interval < minInterval {
		interval = minInterval
	}

	log.Printf("config: interval=%s snmp=%s timeout=%ds retries=%d",
		interval, cfg.SNMPVersion, cfg.SNMPTimeout, cfg.SNMPRetries)

	scanner.SetConfig(cfg)

	tasks, err := api.GetNetworks()
	if err != nil {
		log.Println("API ERROR:", err)
		return interval
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		start := time.Now()

		result, err := scanner.Scan(task.Network)
		if err != nil {
			log.Printf("SCAN ERROR %s: %v", task.Network, err)
			continue
		}

		result.TaskID = task.ID

		log.Printf("scanned %s: %d devices, %d switches in %s",
			task.Network, len(result.Devices), len(result.Switches),
			time.Since(start).Round(time.Millisecond))

		if err := api.Send(result); err != nil {
			log.Println("SEND ERROR:", err)
		}
	}

	return interval
}
