// Package ratelimit ofrece un limitador de ventana fija en memoria, por clave
// (normalmente la IP del cliente), para endpoints públicos sensibles.
package ratelimit

import (
	"os"
	"strconv"
	"sync"
	"time"
)

type entry struct {
	count int
	reset time.Time
}

// Limiter permite como máximo max peticiones por clave en cada ventana.
type Limiter struct {
	mu         sync.Mutex
	m          map[string]*entry
	max        int
	window     time.Duration
	maxEntries int
	now        func() time.Time
}

// New crea un limitador. maxEntries es el tamaño del mapa a partir del cual
// se purgan las ventanas vencidas (el mapa no crece sin límite).
func New(max int, window time.Duration, maxEntries int) *Limiter {
	return &Limiter{m: make(map[string]*entry), max: max, window: window, maxEntries: maxEntries, now: time.Now}
}

// Allow registra una petición de key y indica si está dentro del límite.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.m) >= l.maxEntries {
		for k, e := range l.m {
			if now.After(e.reset) {
				delete(l.m, k)
			}
		}
	}
	e, ok := l.m[key]
	if !ok || now.After(e.reset) {
		l.m[key] = &entry{count: 1, reset: now.Add(l.window)}
		return true
	}
	if e.count >= l.max {
		return false
	}
	e.count++
	return true
}

// EnvInt lee un entero positivo de la variable de entorno k; si falta o no es
// válido devuelve d.
func EnvInt(k string, d int) int {
	if n, err := strconv.Atoi(os.Getenv(k)); err == nil && n > 0 {
		return n
	}
	return d
}
