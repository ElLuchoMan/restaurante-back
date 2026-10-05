package clientip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func req(remote string, xff ...string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remote
	for _, v := range xff {
		r.Header.Add("X-Forwarded-For", v)
	}
	return r
}

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name string
		r    *http.Request
		hops int
		want string
	}{
		{"sin proxy usa RemoteAddr", req("192.0.2.5:80"), 2, "192.0.2.5"},
		{"hops 0 ignora XFF", req("192.0.2.5:80", "6.6.6.6"), 0, "192.0.2.5"},
		{"RemoteAddr sin puerto", req("192.0.2.5"), 0, "192.0.2.5"},
		{"RemoteAddr no IP", req("invalid"), 0, "invalid"},
		{"2 hops: penúltima entrada", req("10.0.0.1:1", "6.6.6.6, 203.0.113.9, 172.70.0.1"), 2, "203.0.113.9"},
		{"1 hop: última entrada", req("10.0.0.1:1", "6.6.6.6, 203.0.113.9"), 1, "203.0.113.9"},
		{"varias cabeceras XFF", req("10.0.0.1:1", "6.6.6.6", "203.0.113.9, 172.70.0.1"), 2, "203.0.113.9"},
		{"menos entradas que hops", req("10.0.0.1:1", "6.6.6.6"), 2, "10.0.0.1"},
		{"entrada no válida", req("10.0.0.1:1", "6.6.6.6, no-es-ip, 172.70.0.1"), 2, "10.0.0.1"},
		{"IPv6 agrupada por /64", req("10.0.0.1:1", "2001:db8:1:2:aaaa:bbbb:cccc:dddd, 172.70.0.1"), 2, "2001:db8:1:2::/64"},
		{"IPv4 mapeada", req("[::ffff:192.0.2.9]:80"), 0, "192.0.2.9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromRequest(tt.r, tt.hops); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHopsFromEnv(t *testing.T) {
	cases := map[string]int{"": DefaultTrustedProxyHops, "0": 0, "1": 1, " 3 ": 3, "10": 10, "11": DefaultTrustedProxyHops, "-1": DefaultTrustedProxyHops, "x": DefaultTrustedProxyHops}
	for v, want := range cases {
		t.Setenv(EnvTrustedProxyHops, v)
		if got := HopsFromEnv(); got != want {
			t.Fatalf("%q: got %d, want %d", v, got, want)
		}
	}
}
