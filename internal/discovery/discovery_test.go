package discovery

import (
	"testing"

	"github.com/miekg/dns"
)

func TestParseHomeyResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *dns.Msg
		want     *HomeyCandidate
	}{
		{
			name: "valid homey response with all fields",
			response: func() *dns.Msg {
				m := new(dns.Msg)
				m.Answer = []dns.RR{
					&dns.PTR{
						Hdr: dns.RR_Header{Name: "_homey._tcp.local.", Rrtype: dns.TypePTR},
						Ptr: "Homey Self-Hosted Server._homey._tcp.local.",
					},
					&dns.TXT{
						Hdr: dns.RR_Header{Name: "Homey Self-Hosted Server._homey._tcp.local.", Rrtype: dns.TypeTXT},
						Txt: []string{"id=abc123", "name=My Homey", "model=shs", "version=12.0.0"},
					},
					&dns.SRV{
						Hdr:    dns.RR_Header{Name: "Homey Self-Hosted Server._homey._tcp.local.", Rrtype: dns.TypeSRV},
						Port:   4859,
						Target: "homey.local.",
					},
					&dns.A{
						Hdr: dns.RR_Header{Name: "homey.local.", Rrtype: dns.TypeA},
						A:   []byte{10, 0, 1, 1},
					},
				}
				return m
			}(),
			want: &HomeyCandidate{
				Address:  "http://10.0.1.1:4859",
				Host:     "10.0.1.1",
				Port:     4859,
				HomeyID:  "abc123",
				Name:     "My Homey",
				Model:    "shs",
				Version:  "12.0.0",
				Instance: "Homey Self-Hosted Server._homey._tcp.local",
			},
		},
		{
			name: "response with hostname only (no A record)",
			response: func() *dns.Msg {
				m := new(dns.Msg)
				m.Answer = []dns.RR{
					&dns.SRV{
						Hdr:    dns.RR_Header{Name: "test._homey._tcp.local.", Rrtype: dns.TypeSRV},
						Port:   443,
						Target: "homey-pro.local.",
					},
				}
				return m
			}(),
			want: &HomeyCandidate{
				Address: "https://homey-pro.local:443",
				Host:    "homey-pro.local",
				Port:    443,
			},
		},
		{
			name: "response without SRV record returns nil",
			response: func() *dns.Msg {
				m := new(dns.Msg)
				m.Answer = []dns.RR{
					&dns.PTR{
						Hdr: dns.RR_Header{Name: "_homey._tcp.local.", Rrtype: dns.TypePTR},
						Ptr: "Test._homey._tcp.local.",
					},
				}
				return m
			}(),
			want: nil,
		},
		{
			name:     "empty response returns nil",
			response: new(dns.Msg),
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHomeyResponse(tt.response)

			if tt.want == nil {
				if got != nil {
					t.Errorf("parseHomeyResponse() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("parseHomeyResponse() = nil, want %v", tt.want)
			}

			if got.Address != tt.want.Address {
				t.Errorf("Address = %q, want %q", got.Address, tt.want.Address)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", got.Host, tt.want.Host)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %d, want %d", got.Port, tt.want.Port)
			}
			if got.HomeyID != tt.want.HomeyID {
				t.Errorf("HomeyID = %q, want %q", got.HomeyID, tt.want.HomeyID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %q, want %q", got.Model, tt.want.Model)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.want.Version)
			}
		})
	}
}

func TestParseHomeyResponse_TXTFields(t *testing.T) {
	// Test that TXT fields are parsed correctly
	m := new(dns.Msg)
	m.Answer = []dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{Rrtype: dns.TypeTXT},
			Txt: []string{
				"id=homey-12345",
				"name=Living Room Homey",
				"model=pro",
				"version=10.5.2",
				"unknown=ignored",
			},
		},
		&dns.SRV{
			Hdr:    dns.RR_Header{Rrtype: dns.TypeSRV},
			Port:   80,
			Target: "test.local.",
		},
	}

	got := parseHomeyResponse(m)
	if got == nil {
		t.Fatal("parseHomeyResponse() returned nil")
	}

	if got.HomeyID != "homey-12345" {
		t.Errorf("HomeyID = %q, want %q", got.HomeyID, "homey-12345")
	}
	if got.Name != "Living Room Homey" {
		t.Errorf("Name = %q, want %q", got.Name, "Living Room Homey")
	}
	if got.Model != "pro" {
		t.Errorf("Model = %q, want %q", got.Model, "pro")
	}
	if got.Version != "10.5.2" {
		t.Errorf("Version = %q, want %q", got.Version, "10.5.2")
	}
}

func TestParseHomeyResponse_IPv6(t *testing.T) {
	m := new(dns.Msg)
	m.Answer = []dns.RR{
		&dns.SRV{
			Hdr:    dns.RR_Header{Rrtype: dns.TypeSRV},
			Port:   4859,
			Target: "homey.local.",
		},
		&dns.AAAA{
			Hdr:  dns.RR_Header{Rrtype: dns.TypeAAAA},
			AAAA: []byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
	}

	got := parseHomeyResponse(m)
	if got == nil {
		t.Fatal("parseHomeyResponse() returned nil")
	}

	// Should use IPv6 address
	if got.Host != "fe80::1" {
		t.Errorf("Host = %q, want %q", got.Host, "fe80::1")
	}
}

func TestParseHomeyResponse_PreferIPv4(t *testing.T) {
	m := new(dns.Msg)
	m.Answer = []dns.RR{
		&dns.SRV{
			Hdr:    dns.RR_Header{Rrtype: dns.TypeSRV},
			Port:   4859,
			Target: "homey.local.",
		},
		&dns.AAAA{
			Hdr:  dns.RR_Header{Rrtype: dns.TypeAAAA},
			AAAA: []byte{0xfe, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		&dns.A{
			Hdr: dns.RR_Header{Rrtype: dns.TypeA},
			A:   []byte{192, 168, 1, 100},
		},
	}

	got := parseHomeyResponse(m)
	if got == nil {
		t.Fatal("parseHomeyResponse() returned nil")
	}

	// Should prefer IPv4 over IPv6
	if got.Host != "192.168.1.100" {
		t.Errorf("Host = %q, want %q", got.Host, "192.168.1.100")
	}
}
