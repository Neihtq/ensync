// Package discovery implements mDNS for the grandmaster to discover the followers
package discovery

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

const mdnsName = "_ensync._tcp"

type DiscoveryService struct {
	Followers *follower.Followers
	NTPPort   string
}

func NewDiscoveryService(followers *follower.Followers, ntpPort string) *DiscoveryService {
	return &DiscoveryService{
		Followers: followers,
		NTPPort:   ntpPort,
	}
}

func (ds *DiscoveryService) Discover() {
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go ds.DiscoverFollower(entriesCh)
	go ds.ScanForServers(entriesCh)
	go ds.ScanForServersMacNative(entriesCh)
}

func (ds *DiscoveryService) ScanForServers(entriesCh chan *mdns.ServiceEntry) {
	for {
		ifaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Error getting interfaces:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("Query for Service:", mdnsName)

		for _, ifc := range ifaces {
			if (ifc.Flags&net.FlagLoopback) != 0 || (ifc.Flags&net.FlagUp) == 0 || (ifc.Flags&net.FlagMulticast) == 0 {
				continue
			}

			i := ifc
			params := &mdns.QueryParam{
				Service:             mdnsName,
				Domain:              "",
				Timeout:             2 * time.Second,
				Interface:           &i,
				Entries:             entriesCh,
				WantUnicastResponse: false,
				DisableIPv6:         true,
			}

			go func(p *mdns.QueryParam) {
				_ = mdns.Query(p)
			}(params)
		}

		time.Sleep(2 * time.Second)
	}
}

func (ds *DiscoveryService) DiscoverFollower(entriesCh chan *mdns.ServiceEntry) {
	for entry := range entriesCh {
		if entry == nil || entry.AddrV4 == nil || len(entry.InfoFields) == 0 {
			continue
		}

		if !strings.Contains(entry.Name, "_ensync") {
			continue
		}

		endpoint := entry.InfoFields[0]
		ipAddress := entry.AddrV4.String()
		url := ipAddress + ":" + strconv.Itoa(entry.Port) + endpoint
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		fmt.Println("[Discovery] Found entry ", ipAddress, endpoint, entry.Port, entry.Name)
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		if _, exists := ds.Followers.Followers[ipAddress]; !exists {
			follower.SubscribeFollower(ds.Followers, url, ds.NTPPort)
		}
	}
}

func (ds *DiscoveryService) ScanForServersMacNative(entriesCh chan *mdns.ServiceEntry) {
	if runtime.GOOS != "darwin" {
		return
	}

	for {
		cmd := exec.Command("dns-sd", "-B", mdnsName, "local.")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if err := cmd.Start(); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, mdnsName) && strings.Contains(line, "Add") {
					fields := strings.Fields(line)
					if len(fields) >= 7 {
						instance := strings.Join(fields[6:], " ")

						// Resolve IP
						resolveCmd := exec.Command("sh", "-c", fmt.Sprintf("dns-sd -G v4 '%s' | grep -m 1 Add", instance+".local."))
						out, resolveErr := resolveCmd.Output()
						if resolveErr == nil && len(out) > 0 {
							resolveFields := strings.Fields(string(out))
							if len(resolveFields) >= 6 {
								ipStr := resolveFields[5]

								// Resolve Port using -L
								resolvePortCmd := exec.Command("sh", "-c", fmt.Sprintf("dns-sd -L '%s' %s | grep 'can be reached at'", instance, mdnsName))
								portOut, portErr := resolvePortCmd.Output()
								port := 9001
								if portErr == nil && len(portOut) > 0 {
									// Extract port
									portOutStr := string(portOut)
									portParts := strings.Split(portOutStr, ":")
									if len(portParts) > 1 {
										portParse := strings.Fields(portParts[len(portParts)-1])
										if len(portParse) > 0 {
											if p, err := strconv.Atoi(portParse[0]); err == nil {
												port = p
											}
										}
									}
								}

								ipAddr := net.ParseIP(ipStr)
								if ipAddr != nil && ipAddr.To4() != nil {
									entry := &mdns.ServiceEntry{
										Name:       fmt.Sprintf("%s.%s.local.", instance, mdnsName),
										Host:       fmt.Sprintf("%s.local.", instance),
										AddrV4:     ipAddr,
										Port:       port,
										InfoFields: []string{"/connections"},
									}

									// Sending in non-blocking way
									select {
									case entriesCh <- entry:
									default:
									}
								}
							}
						}
					}
				}
			}
		}()

		err = cmd.Wait()
		if err != nil {
			time.Sleep(2 * time.Second)
		}
	}
}
