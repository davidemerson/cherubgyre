package services

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPLimiterAllowsUnderLimit(t *testing.T) {
	l := NewIPLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if err := l.Allow("1.2.3.4"); err != nil {
			t.Fatalf("attempt %d should be allowed, got %v", i+1, err)
		}
	}
}

func TestIPLimiterBlocksAtLimit(t *testing.T) {
	l := NewIPLimiter(2, time.Minute)
	_ = l.Allow("1.2.3.4")
	_ = l.Allow("1.2.3.4")
	if err := l.Allow("1.2.3.4"); err == nil {
		t.Fatal("3rd attempt should be rate-limited, got nil error")
	}
}

func TestIPLimiterIsolatesIPs(t *testing.T) {
	l := NewIPLimiter(1, time.Minute)
	if err := l.Allow("1.1.1.1"); err != nil {
		t.Fatalf("1.1.1.1 first attempt should be allowed: %v", err)
	}
	if err := l.Allow("2.2.2.2"); err != nil {
		t.Fatalf("2.2.2.2 first attempt should be allowed (different IP): %v", err)
	}
	if err := l.Allow("1.1.1.1"); err == nil {
		t.Fatal("1.1.1.1 second attempt should be blocked")
	}
}

func TestIPLimiterWindowExpiryResetsCounter(t *testing.T) {
	// Very short window so we can observe expiry in a unit test.
	l := NewIPLimiter(1, 50*time.Millisecond)
	if err := l.Allow("1.1.1.1"); err != nil {
		t.Fatalf("first attempt should be allowed: %v", err)
	}
	if err := l.Allow("1.1.1.1"); err == nil {
		t.Fatal("second attempt within window should be blocked")
	}
	time.Sleep(75 * time.Millisecond)
	if err := l.Allow("1.1.1.1"); err != nil {
		t.Fatalf("after window expiry, attempt should be allowed: %v", err)
	}
}

func TestClientIPFallbackChain(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xreal      string
		want       string
	}{
		{
			name:       "remote addr only",
			remoteAddr: "192.0.2.10:54321",
			want:       "192.0.2.10",
		},
		{
			name:  "x-real-ip wins over remote addr",
			xreal: "203.0.113.5",
			want:  "203.0.113.5",
		},
		{
			name: "x-forwarded-for wins over x-real-ip",
			xff:  "198.51.100.1",
			want: "198.51.100.1",
		},
		{
			name: "x-forwarded-for takes first entry",
			xff:  "198.51.100.1, 10.0.0.2, 10.0.0.3",
			want: "198.51.100.1",
		},
		{
			name: "x-forwarded-for whitespace stripped",
			xff:  "   198.51.100.2  ",
			want: "198.51.100.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tt.remoteAddr != "" {
				r.RemoteAddr = tt.remoteAddr
			}
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xreal != "" {
				r.Header.Set("X-Real-IP", tt.xreal)
			}
			if got := ClientIP(r); got != tt.want {
				t.Errorf("ClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
