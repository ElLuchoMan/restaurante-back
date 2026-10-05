// Package clientip determina la IP real del cliente detrás de proxies de
// confianza sin dejar que el cliente la falsifique con X-Forwarded-For.
//
// Cada proxy de confianza AÑADE al final de X-Forwarded-For la IP desde la que
// recibió la conexión; los valores anteriores los controla el cliente y no son
// fiables. Con N proxies de confianza delante de la app, la IP del cliente es
// la N-ésima entrada contando desde el final.
package clientip

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// EnvTrustedProxyHops es la variable de entorno con el número de proxies de
// confianza que añaden una entrada a X-Forwarded-For.
const EnvTrustedProxyHops = "TRUSTED_PROXY_HOPS"

// DefaultTrustedProxyHops es el valor para Render: Cloudflare (borde de Render)
// añade la IP del cliente y el proxy de Render añade la IP de Cloudflare, de
// modo que el cliente es la penúltima entrada (la N=2 desde el final).
const DefaultTrustedProxyHops = 2

// maxHops acota el valor configurable para evitar configuraciones absurdas.
const maxHops = 10

// HopsFromEnv lee TRUSTED_PROXY_HOPS (entero entre 0 y 10; 0 = sin proxy, se
// ignora X-Forwarded-For). Si falta o no es válido devuelve el valor por defecto.
func HopsFromEnv() int {
	n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(EnvTrustedProxyHops)))
	if err != nil || n < 0 || n > maxHops {
		return DefaultTrustedProxyHops
	}
	return n
}

// FromRequest devuelve la IP del cliente normalizada. Con hops > 0 usa la
// entrada hops-ésima desde el final de X-Forwarded-For; si la cabecera no
// existe, tiene menos entradas que hops o esa entrada no es una IP válida,
// recurre a la IP de la conexión (RemoteAddr). Las IPv6 se agrupan por /64.
func FromRequest(r *http.Request, hops int) string {
	if hops > 0 {
		parts := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
		if len(parts) >= hops {
			if ip := net.ParseIP(strings.TrimSpace(parts[len(parts)-hops])); ip != nil {
				return normalize(ip)
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return normalize(ip)
	}
	return host
}

func normalize(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}
