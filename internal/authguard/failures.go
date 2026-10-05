// Package authguard limita los fallos de autenticación por clave (documento)
// con espera creciente, en memoria y de tamaño acotado.
package authguard

import (
	"sync"
	"time"
)

// Config parametriza Failures.
type Config struct {
	MaxFailures int           // fallos permitidos antes de bloquear
	Window      time.Duration // ventana de conteo, desde el primer fallo
	BaseWait    time.Duration // espera tras alcanzar MaxFailures; se duplica en cada fallo posterior
	MaxWait     time.Duration // tope de la espera
	MaxEntries  int           // tamaño del mapa a partir del cual se purgan entradas vencidas
}

type entry struct {
	count        int
	reset        time.Time
	blockedUntil time.Time
}

// Failures cuenta fallos por clave. Es seguro para uso concurrente.
//
// Trade-off: el bloqueo es por documento, así que un atacante puede impedir
// temporalmente el login de un tercero fallando a propósito. Se mitiga con
// bloqueos temporales (MaxWait) y porque la IP del atacante también está
// limitada; nunca es indefinido.
type Failures struct {
	mu  sync.Mutex
	m   map[int64]*entry
	cfg Config
	now func() time.Time
}

// NewFailures crea el contador.
func NewFailures(cfg Config) *Failures {
	return &Failures{m: make(map[int64]*entry), cfg: cfg, now: time.Now}
}

// Check indica si la clave puede intentar autenticarse; si no, devuelve cuánto
// falta para el siguiente intento.
func (f *Failures) Check(key int64) (bool, time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.m[key]
	if !ok {
		return true, 0
	}
	if wait := e.blockedUntil.Sub(f.now()); wait > 0 {
		return false, wait
	}
	return true, 0
}

// Fail registra un fallo de autenticación.
func (f *Failures) Fail(key int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := f.now()
	e, ok := f.m[key]
	if !ok || f.expired(e, now) {
		if !ok && len(f.m) >= f.cfg.MaxEntries {
			f.purge(now)
			if len(f.m) >= f.cfg.MaxEntries {
				return // mapa lleno de entradas vigentes: no se amplía (memoria acotada)
			}
		}
		e = &entry{reset: now.Add(f.cfg.Window)}
		f.m[key] = e
	}
	e.count++
	if e.count >= f.cfg.MaxFailures {
		e.blockedUntil = now.Add(f.wait(e.count))
	}
}

// Reset borra el historial de la clave (login correcto).
func (f *Failures) Reset(key int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
}

func (f *Failures) wait(count int) time.Duration {
	d := f.cfg.BaseWait
	for i := f.cfg.MaxFailures; i < count && d < f.cfg.MaxWait; i++ {
		d *= 2
	}
	if d > f.cfg.MaxWait {
		return f.cfg.MaxWait
	}
	return d
}

func (f *Failures) expired(e *entry, now time.Time) bool {
	return now.After(e.reset) && !now.Before(e.blockedUntil)
}

func (f *Failures) purge(now time.Time) {
	for k, e := range f.m {
		if f.expired(e, now) {
			delete(f.m, k)
		}
	}
}

// RetryAfterSeconds convierte una espera en segundos enteros (mínimo 1) para
// la cabecera Retry-After.
func RetryAfterSeconds(d time.Duration) int {
	s := int((d + time.Second - 1) / time.Second)
	if s < 1 {
		return 1
	}
	return s
}
