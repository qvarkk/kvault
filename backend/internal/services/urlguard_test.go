package services

import (
	"net"
	"testing"
)

func TestValidatePublicURL(t *testing.T) {
	cases := []struct {
		raw     string
		wantErr bool
	}{
		{"https://example.com/article", false},
		{"http://example.com", false},
		{"ftp://example.com/file", true},
		{"file:///etc/passwd", true},
		{"https://", true},
		{"not a url", true},
		{"", true},
	}
	for _, c := range cases {
		_, err := validatePublicURL(c.raw)
		if (err != nil) != c.wantErr {
			t.Errorf("validatePublicURL(%q) err=%v, wantErr=%v", c.raw, err, c.wantErr)
		}
	}
}

func TestIsDisallowedIP(t *testing.T) {
	cases := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},       // loopback
		{"::1", true},             // loopback v6
		{"169.254.169.254", true}, // link-local (cloud metadata)
		{"10.0.0.1", true},        // private
		{"192.168.1.1", true},     // private
		{"172.16.0.1", true},      // private
		{"100.64.0.1", true},      // CGNAT
		{"0.0.0.0", true},         // unspecified
		{"fd00::1", true},         // ULA
		{"8.8.8.8", false},        // public
		{"1.1.1.1", false},        // public
	}
	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		if ip == nil {
			t.Fatalf("bad test ip %q", c.ip)
		}
		if got := isDisallowedIP(ip); got != c.blocked {
			t.Errorf("isDisallowedIP(%s) = %v, want %v", c.ip, got, c.blocked)
		}
	}
	if !isDisallowedIP(nil) {
		t.Error("nil IP must be treated as disallowed")
	}
}
