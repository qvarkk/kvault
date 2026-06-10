package services

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

// cgnatRange is the carrier-grade NAT block 100.64.0.0/10 (RFC 6598), which
// net.IP.IsPrivate does not cover.
var cgnatRange = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// validatePublicURL parses raw and enforces an http(s) scheme and a non-empty
// host. It does not resolve DNS — IP-level blocking happens at dial time so that
// DNS rebinding cannot bypass it.
func validatePublicURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, errors.New("url has no host")
	}
	return u, nil
}

// isDisallowedIP reports whether an address must not be reached from the server:
// loopback, private (RFC1918 + IPv6 ULA), link-local, unspecified, or CGNAT.
func isDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil && cgnatRange.Contains(v4) {
		return true
	}
	return false
}

// newGuardedHTTPClient returns an http.Client that refuses to connect to
// disallowed IP ranges (checked at dial time, after DNS resolution, so rebinding
// is ineffective) and caps redirects, re-validating every hop.
func newGuardedHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if isDisallowedIP(ip) {
				return fmt.Errorf("blocked connection to disallowed address %s", host)
			}
			return nil
		},
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			if _, err := validatePublicURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
	}
}

const (
	maxRedirects      = 5
	maxFetchBodyBytes = 5 << 20 // 5 MiB cap on fetched URL bodies
)
