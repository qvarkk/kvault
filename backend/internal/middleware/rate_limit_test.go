package middleware

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiterBurstThenDeny(t *testing.T) {
	// One token per hour, burst of 3 → first 3 allowed, 4th denied.
	l := newIPRateLimiter(rate.Every(time.Hour), 3)

	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4th request should be denied after burst is exhausted")
	}
}

func TestIPRateLimiterPerIPIndependent(t *testing.T) {
	l := newIPRateLimiter(rate.Every(time.Hour), 1)

	if !l.allow("10.0.0.1") {
		t.Fatal("first IP first request should be allowed")
	}
	if l.allow("10.0.0.1") {
		t.Fatal("first IP second request should be denied")
	}
	if !l.allow("10.0.0.2") {
		t.Fatal("second IP should have its own independent budget")
	}
}
