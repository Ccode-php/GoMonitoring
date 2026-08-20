package scanner

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"GoMonitoring/models"

	"github.com/gosnmp/gosnmp"
)

const (

	// Switch hostname
	oidSysName = ".1.3.6.1.2.1.1.5.0"

	// Port nomlari (2-qismda ishlatiladi)
	oidIfName = ".1.3.6.1.2.1.31.1.1.1.1"

	// MAC Table (3-qismda ishlatiladi)
	oidBridgeMac = ".1.3.6.1.2.1.17.4.3.1.1"

	// MAC -> Bridge Port (3-qismda ishlatiladi)
	oidBridgePort = ".1.3.6.1.2.1.17.4.3.1.2"

	// Bridge Port -> IfIndex (3-qismda ishlatiladi)
	oidBasePortIfIndex = ".1.3.6.1.2.1.17.1.4.1.2"
)

func connect(ip string) (*gosnmp.GoSNMP, error) {

	cfg := GetConfig()

	version := gosnmp.Version2c

	switch cfg.SNMPVersion {

	case "v1":
		version = gosnmp.Version1

	case "v2c":
		version = gosnmp.Version2c

	default:
		version = gosnmp.Version2c
	}

	timeout := cfg.SNMPTimeout

	if timeout <= 0 {
		timeout = 2
	}

	retries := cfg.SNMPRetries

	if retries < 0 {
		retries = 1
	}

	community := cfg.SNMPCommunity

	if community == "" {
		community = "public"
	}

	g := &gosnmp.GoSNMP{

		Target: ip,

		Port: 161,

		Community: community,

		Version: version,

		Timeout: time.Duration(
			timeout,
		) * time.Second,

		Retries: retries,
	}

	err := g.Connect()

	if err != nil {
		return nil, err
	}

	return g, nil
}

func getSysName(g *gosnmp.GoSNMP) string {

	result, err := g.Get([]string{
		oidSysName,
	})

	if err != nil {

		return ""

	}

	if len(result.Variables) == 0 {

		return ""

	}

	value, ok := result.Variables[0].Value.(string)

	if !ok {

		return ""

	}

	return strings.TrimSpace(value)
}

func GetSwitchInfo(ip string) (models.Switch, bool) {

	var sw models.Switch

	g, err := connect(ip)

	if err != nil {

		return sw, false

	}

	defer g.Conn.Close()

	sw.IP = ip

	sw.Hostname = getSysName(g)

	if sw.Hostname == "" {

		return sw, false

	}

	fmt.Println("SNMP:", sw.Hostname)

	sw.Ports = getPorts(g)

	bridgeMap := getBridgePortMap(g)

	sw.Ports = getMacTable(
		g,
		bridgeMap,
		sw.Ports,
	)

	return sw, true
}

func getPorts(g *gosnmp.GoSNMP) []models.Port {

	var ports []models.Port

	err := g.Walk(
		oidIfName,
		func(pdu gosnmp.SnmpPDU) error {

			name, ok := pdu.Value.(string)

			if !ok {
				return nil
			}

			parts := strings.Split(pdu.Name, ".")

			indexText := parts[len(parts)-1]

			index, err := strconv.Atoi(indexText)

			if err != nil {
				return nil
			}

			port := models.Port{
				Index: index,
				Name:  name,
				MACs:  []string{},
			}

			ports = append(ports, port)

			return nil
		},
	)

	if err != nil {
		return []models.Port{}
	}

	return ports
}

func getBridgePortMap(g *gosnmp.GoSNMP) map[int]int {

	mapping := make(map[int]int)

	err := g.Walk(
		oidBasePortIfIndex,
		func(pdu gosnmp.SnmpPDU) error {

			parts := strings.Split(pdu.Name, ".")

			bridgePort, err := strconv.Atoi(parts[len(parts)-1])

			if err != nil {
				return nil
			}

			var ifIndex int

			switch v := pdu.Value.(type) {

			case int:
				ifIndex = v

			case uint:
				ifIndex = int(v)

			case uint64:
				ifIndex = int(v)

			case int64:
				ifIndex = int(v)

			default:
				return nil

			}

			mapping[bridgePort] = ifIndex

			return nil

		},
	)

	if err != nil {

		return map[int]int{}

	}

	return mapping

}

func getMacTable(
	g *gosnmp.GoSNMP,
	bridgeMap map[int]int,
	ports []models.Port,
) []models.Port {

	err := g.Walk(
		oidBridgePort,
		func(pdu gosnmp.SnmpPDU) error {

			var bridgePort int

			switch v := pdu.Value.(type) {

			case int:
				bridgePort = v

			case uint:
				bridgePort = int(v)

			case uint64:
				bridgePort = int(v)

			case int64:
				bridgePort = int(v)

			default:
				return nil
			}

			ifIndex := bridgeMap[bridgePort]

			oid := strings.TrimPrefix(
				pdu.Name,
				oidBridgePort+".",
			)

			macOID := strings.Split(
				oid,
				".",
			)

			if len(macOID) != 6 {

				return nil

			}

			mac := fmt.Sprintf(
				"%02X:%02X:%02X:%02X:%02X:%02X",
				toInt(macOID[0]),
				toInt(macOID[1]),
				toInt(macOID[2]),
				toInt(macOID[3]),
				toInt(macOID[4]),
				toInt(macOID[5]),
			)

			for i := range ports {

				if ports[i].Index == ifIndex {

					ports[i].MACs = append(
						ports[i].MACs,
						mac,
					)

				}

			}

			return nil

		},
	)

	if err != nil {

		return ports

	}

	return ports

}

func toInt(s string) int {

	n, _ := strconv.Atoi(s)

	return n

}
