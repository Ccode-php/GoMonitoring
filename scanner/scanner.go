package scanner

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"GoMonitoring/models"
)

const (
	// Bir vaqtda ishlaydigan eng ko'p hostlar soni
	maxWorkers = 64

	// Eng katta ruxsat etilgan tarmoq: /16 (65534 host)
	maxHostBits = 16
)

// hostsOf CIDR'dan (masalan 192.168.1.0/24 yoki 10.0.0.0/22) barcha host IP'larni qaytaradi.
func hostsOf(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return nil, fmt.Errorf("noto'g'ri tarmoq %q: %w", cidr, err)
	}

	if ip.To4() == nil {
		return nil, fmt.Errorf("%q: faqat IPv4 qo'llab-quvvatlanadi", cidr)
	}

	ones, bits := ipnet.Mask.Size()
	if bits-ones > maxHostBits {
		return nil, fmt.Errorf("%q juda katta (eng katta /%d)", cidr, bits-maxHostBits)
	}

	var hosts []string
	for cur := ipnet.IP.To4(); ipnet.Contains(cur); cur = nextIP(cur) {
		hosts = append(hosts, cur.String())
	}

	// Tarmoq va broadcast manzillarini olib tashlaymiz (/31, /32 dan tashqari)
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}

	return hosts, nil
}

func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}

	return next
}

// ping: bitta ICMP so'rov
func ping(ip string) bool {
	return exec.Command("ping", "-c", "1", "-W", "1", ip).Run() == nil
}

// neighbor: ARP jadvalidan MAC va holatni oladi (REACHABLE, STALE, FAILED...)
func neighbor(ip string) (mac, state string) {
	out, err := exec.Command("ip", "neigh", "show", ip).Output()
	if err != nil {
		return "", ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)

		for i := 0; i+1 < len(f); i++ {
			if f[i] == "lladdr" {
				return strings.ToUpper(f[i+1]), f[len(f)-1]
			}
		}
	}

	return "", ""
}

// getHostname: DNS so'rovi timeout bilan
func getHostname(ip string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}

	return strings.TrimSuffix(names[0], ".")
}

// scanHost: bitta hostni tekshiradi.
// Host ICMP'ga javob bersa yoki (ICMP yopiq bo'lsa ham) ARP'ga javob bergan bo'lsa tirik hisoblanadi.
func scanHost(ip string) *models.Device {
	pingOK := ping(ip)

	mac, state := neighbor(ip)

	if !pingOK && state != "REACHABLE" {
		return nil
	}

	// MAC yo'q bo'lsa (boshqa subnet yoki scanner'ning o'zi) qurilmani o'tkazib yuboramiz
	if mac == "" {
		return nil
	}

	return &models.Device{
		IP:       ip,
		MAC:      mac,
		Hostname: getHostname(ip),
	}
}

// Scan tarmoqni worker pool bilan tekshiradi.
func Scan(network string) (models.ScanResult, error) {
	result := models.ScanResult{
		Devices:  make([]models.Device, 0),
		Switches: make([]models.Switch, 0),
	}

	hosts, err := hostsOf(network)
	if err != nil {
		return result, err
	}

	workers := maxWorkers
	if len(hosts) < workers {
		workers = len(hosts)
	}

	jobs := make(chan string)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for w := 0; w < workers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for ip := range jobs {
				device := scanHost(ip)
				if device == nil {
					continue
				}

				sw, isSwitch := GetSwitchInfo(ip)
				device.IsSwitch = isSwitch

				mu.Lock()
				result.Devices = append(result.Devices, *device)
				if isSwitch {
					result.Switches = append(result.Switches, sw)
				}
				mu.Unlock()

				log.Printf("FOUND %-15s %s switch=%v", device.IP, device.MAC, isSwitch)
			}
		}()
	}

	for _, ip := range hosts {
		jobs <- ip
	}
	close(jobs)

	wg.Wait()

	return result, nil
}
