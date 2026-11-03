package scanner

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"GoMonitoring/models"

	"github.com/gosnmp/gosnmp"
)

const (
	oidSysName         = ".1.3.6.1.2.1.1.5.0"
	oidIfName          = ".1.3.6.1.2.1.31.1.1.1.1"
	oidBridgeAddress   = ".1.3.6.1.2.1.17.1.1.0"
	oidBridgePort      = ".1.3.6.1.2.1.17.4.3.1.2"
	oidBasePortIfIndex = ".1.3.6.1.2.1.17.1.4.1.2"
)

var warnV3 sync.Once

// ---------------------------------------------------------------------------
// Yordamchi funksiyalar (oldin 2-3 joyda takrorlangan edi)
// ---------------------------------------------------------------------------

func formatMAC(b []byte) string {
	if len(b) != 6 {
		return ""
	}

	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		b[0], b[1], b[2], b[3], b[4], b[5])
}

// gosnmp OctetString'ni []byte yoki string qilib qaytarishi mumkin
func pduString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []byte:
		return strings.TrimSpace(string(t))
	}

	return ""
}

func pduBytes(v interface{}) []byte {
	switch t := v.(type) {
	case []byte:
		return t
	case string:
		return []byte(t)
	}

	return nil
}

func pduInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint:
		return int(t), true
	case uint32:
		return int(t), true
	case uint64:
		return int(t), true
	}

	return 0, false
}

// OID oxiridagi son (".1.3.6...5" -> 5)
func lastIndex(oid string) (int, bool) {
	i := strings.LastIndex(oid, ".")
	n, err := strconv.Atoi(oid[i+1:])

	return n, err == nil
}

// ---------------------------------------------------------------------------
// SNMP ulanish
// ---------------------------------------------------------------------------

func connect(ip string) (*gosnmp.GoSNMP, error) {
	cfg := GetConfig()

	var version gosnmp.SnmpVersion

	switch cfg.SNMPVersion {
	case "v1":
		version = gosnmp.Version1
	case "v2c", "":
		version = gosnmp.Version2c
	default:
		warnV3.Do(func() {
			log.Printf("SNMP %q qo'llab-quvvatlanmaydi, faqat v1 va v2c. Switchlar tekshirilmaydi.", cfg.SNMPVersion)
		})

		return nil, fmt.Errorf("unsupported snmp version %q", cfg.SNMPVersion)
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
		Target:    ip,
		Port:      161,
		Community: community,
		Version:   version,
		Timeout:   time.Duration(timeout) * time.Second,
		Retries:   retries,
	}

	if err := g.Connect(); err != nil {
		return nil, err
	}

	return g, nil
}

func getSysName(g *gosnmp.GoSNMP) string {
	result, err := g.Get([]string{oidSysName})
	if err != nil || len(result.Variables) == 0 {
		return ""
	}

	return pduString(result.Variables[0].Value)
}

func getBridgeAddress(g *gosnmp.GoSNMP) string {
	result, err := g.Get([]string{oidBridgeAddress})
	if err != nil || len(result.Variables) == 0 {
		return ""
	}

	return formatMAC(pduBytes(result.Variables[0].Value))
}

// GetSwitchInfo hostga SNMP orqali murojaat qiladi. Javob bermasa switch emas.
func GetSwitchInfo(ip string) (models.Switch, bool) {
	var sw models.Switch

	g, err := connect(ip)
	if err != nil {
		return sw, false
	}
	defer g.Conn.Close()

	// SNMP yo'q hostlarda vaqt yo'qotmaslik uchun birinchi so'rov retry'siz
	retries := g.Retries
	g.Retries = 0

	sw.Hostname = getSysName(g)
	if sw.Hostname == "" {
		return sw, false
	}

	g.Retries = retries

	sw.IP = ip
	sw.MAC = getBridgeAddress(g)

	if sw.MAC == "" {
		log.Printf("SNMP %s (%s): bridge MAC olinmadi, Laravel bu switchni o'tkazib yuboradi", ip, sw.Hostname)
	}

	log.Printf("SNMP SWITCH: IP=%s HOSTNAME=%s MAC=%s", sw.IP, sw.Hostname, sw.MAC)

	sw.Ports = getPorts(g)
	sw.Ports = getMacTable(g, getBridgePortMap(g), sw.Ports)

	return sw, true
}

func getPorts(g *gosnmp.GoSNMP) []models.Port {
	ports := []models.Port{}

	err := g.Walk(oidIfName, func(pdu gosnmp.SnmpPDU) error {
		name := pduString(pdu.Value)

		index, ok := lastIndex(pdu.Name)
		if !ok || name == "" {
			return nil
		}

		ports = append(ports, models.Port{
			Index: index,
			Name:  name,
			MACs:  []string{},
		})

		return nil
	})

	if err != nil {
		return []models.Port{}
	}

	return ports
}

// bridge port -> ifIndex
func getBridgePortMap(g *gosnmp.GoSNMP) map[int]int {
	mapping := make(map[int]int)

	_ = g.Walk(oidBasePortIfIndex, func(pdu gosnmp.SnmpPDU) error {
		bridgePort, ok := lastIndex(pdu.Name)
		if !ok {
			return nil
		}

		if ifIndex, ok := pduInt(pdu.Value); ok {
			mapping[bridgePort] = ifIndex
		}

		return nil
	})

	return mapping
}

// MAC jadvalini portlarga taqsimlaydi
func getMacTable(g *gosnmp.GoSNMP, bridgeMap map[int]int, ports []models.Port) []models.Port {
	// ifIndex -> ports slice'idagi o'rni (har MAC uchun butun ro'yxatni aylanmaslik uchun)
	position := make(map[int]int, len(ports))
	for i := range ports {
		position[ports[i].Index] = i
	}

	_ = g.Walk(oidBridgePort, func(pdu gosnmp.SnmpPDU) error {
		bridgePort, ok := pduInt(pdu.Value)
		if !ok {
			return nil
		}

		ifIndex, ok := bridgeMap[bridgePort]
		if !ok {
			return nil
		}

		pos, ok := position[ifIndex]
		if !ok {
			return nil
		}

		parts := strings.Split(strings.TrimPrefix(pdu.Name, oidBridgePort+"."), ".")
		if len(parts) != 6 {
			return nil
		}

		raw := make([]byte, 6)
		for i, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n > 255 {
				return nil
			}

			raw[i] = byte(n)
		}

		ports[pos].MACs = append(ports[pos].MACs, formatMAC(raw))

		return nil
	})

	return ports
}
