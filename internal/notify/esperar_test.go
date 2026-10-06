package notify

// esperar bloquea hasta que terminen los envíos lanzados (solo para tests).
func (a *asyncNotifier) esperar() { a.wg.Wait() }
