package notify

import (
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// asyncNotifier envía en segundo plano. Si el push no está configurado no hace
// nada (ni siquiera lanza una goroutine). Cada evento tiene un tiempo límite
// total; pasado ese tiempo se abandona (el envío en curso termina por su cuenta).
type asyncNotifier struct {
	timeout time.Duration
	wg      sync.WaitGroup
}

// Notificar lanza el envío y regresa de inmediato.
func (a *asyncNotifier) Notificar(ev Evento) {
	if !configurado() {
		return
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					logs.Error("[Notify] Pánico al enviar la notificación (%s): %v", ev.Tipo, r)
				}
			}()
			procesar(ev)
		}()
		timer := time.NewTimer(a.timeout)
		defer timer.Stop()
		select {
		case <-done:
		case <-timer.C:
			logs.Warn("[Notify] Tiempo agotado al enviar la notificación (%s)", ev.Tipo)
		}
	}()
}

// esperar bloquea hasta que terminen los envíos lanzados (solo para tests).
func (a *asyncNotifier) esperar() { a.wg.Wait() }
