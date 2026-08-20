package scanner

import (
	"sync"

	"GoMonitoring/api"
)

var (
	configMu sync.RWMutex

	config api.ScannerConfig = api.ScannerConfig{
		ScanInterval:  60,
		SNMPCommunity: "public",
		SNMPVersion:   "v2c",
		SNMPTimeout:   2,
		SNMPRetries:   1,
	}
)

func SetConfig(c api.ScannerConfig) {

	configMu.Lock()

	config = c

	configMu.Unlock()
}

func GetConfig() api.ScannerConfig {

	configMu.RLock()

	defer configMu.RUnlock()

	return config
}
