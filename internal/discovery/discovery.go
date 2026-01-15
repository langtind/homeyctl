package discovery

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// HomeyCandidate represents a discovered Homey device
type HomeyCandidate struct {
	Address  string
	Host     string
	Port     int
	HomeyID  string
	Name     string
	Model    string
	Version  string
	Instance string
}

// mDNS multicast address
var mdnsAddr = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}

// DiscoverHomeys searches for Homey devices on the local network via mDNS
func DiscoverHomeys(ctx context.Context, timeout time.Duration) ([]HomeyCandidate, error) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Services to query
	services := []string{
		"_homey._tcp.local.",
		"_athom._tcp.local.",
	}

	var mu sync.Mutex
	var candidates []HomeyCandidate

	// Create UDP socket for mDNS
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS socket: %w", err)
	}
	defer conn.Close()

	// Set read deadline based on timeout
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)

	// Send queries for each service type
	for _, service := range services {
		m := new(dns.Msg)
		m.SetQuestion(service, dns.TypePTR)
		m.RecursionDesired = false

		data, err := m.Pack()
		if err != nil {
			continue
		}

		_, err = conn.WriteToUDP(data, mdnsAddr)
		if err != nil {
			continue
		}
	}

	// Read responses until timeout
	buf := make([]byte, 65535)
	for {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			continue
		}

		resp := new(dns.Msg)
		if err := resp.Unpack(buf[:n]); err != nil {
			continue
		}

		// Parse response into candidate
		candidate := parseHomeyResponse(resp)
		if candidate != nil {
			mu.Lock()
			candidates = append(candidates, *candidate)
			mu.Unlock()
		}
	}

done:
	mu.Lock()
	defer mu.Unlock()

	// Deduplicate by address
	seen := make(map[string]bool)
	var unique []HomeyCandidate
	for _, c := range candidates {
		if !seen[c.Address] {
			seen[c.Address] = true
			unique = append(unique, c)
		}
	}

	return unique, nil
}

// parseHomeyResponse extracts Homey info from mDNS response
func parseHomeyResponse(resp *dns.Msg) *HomeyCandidate {
	var candidate HomeyCandidate
	var host string
	var port uint16
	var ip net.IP

	// Combine all records for parsing
	allRecords := append(resp.Answer, resp.Extra...)

	for _, rr := range allRecords {
		switch r := rr.(type) {
		case *dns.PTR:
			candidate.Instance = strings.TrimSuffix(r.Ptr, ".")
		case *dns.TXT:
			for _, txt := range r.Txt {
				if strings.HasPrefix(txt, "id=") {
					candidate.HomeyID = strings.TrimPrefix(txt, "id=")
				} else if strings.HasPrefix(txt, "name=") {
					candidate.Name = strings.TrimPrefix(txt, "name=")
				} else if strings.HasPrefix(txt, "model=") {
					candidate.Model = strings.TrimPrefix(txt, "model=")
				} else if strings.HasPrefix(txt, "version=") {
					candidate.Version = strings.TrimPrefix(txt, "version=")
				}
			}
		case *dns.SRV:
			port = r.Port
			host = strings.TrimSuffix(r.Target, ".")
		case *dns.A:
			ip = r.A
		case *dns.AAAA:
			if ip == nil {
				ip = r.AAAA
			}
		}
	}

	// Need at least host/IP and port
	if port == 0 {
		return nil
	}

	// Prefer IP over hostname for address
	if ip != nil {
		host = ip.String()
	}
	if host == "" {
		return nil
	}

	candidate.Host = host
	candidate.Port = int(port)

	// Build URL
	scheme := "http"
	if port == 443 {
		scheme = "https"
	}
	candidate.Address = fmt.Sprintf("%s://%s:%d", scheme, host, port)

	return &candidate
}

// VerifyHomey checks if an address is a valid Homey by pinging it
func VerifyHomey(ctx context.Context, address string, timeout time.Duration) (string, bool) {
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	// Try the ping endpoint
	url := address + "/api/manager/system/ping"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", false
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	// Get Homey ID from header
	homeyID := resp.Header.Get("X-Homey-ID")
	return homeyID, homeyID != ""
}

// DiscoverAndVerify discovers Homeys and verifies they respond
func DiscoverAndVerify(ctx context.Context, timeout time.Duration) ([]HomeyCandidate, error) {
	candidates, err := DiscoverHomeys(ctx, timeout)
	if err != nil {
		return nil, err
	}

	var verified []HomeyCandidate
	for _, c := range candidates {
		homeyID, ok := VerifyHomey(ctx, c.Address, 2*time.Second)
		if ok {
			// Use verified HomeyID if we got one from ping
			if homeyID != "" {
				c.HomeyID = homeyID
			}
			verified = append(verified, c)
		}
	}

	return verified, nil
}
