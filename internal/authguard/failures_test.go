package authguard

import (
	"sync"
	"testing"
	"time"
)

func newTest(max int) (*Failures, *time.Time) {
	now := time.Unix(1000000, 0)
	f := NewFailures(Config{MaxFailures: 3, Window: 15 * time.Minute, BaseWait: 30 * time.Second, MaxWait: 2 * time.Minute, MaxEntries: max})
	f.now = func() time.Time { return now }
	return f, &now
}

func TestBloqueoCrecienteYTope(t *testing.T) {
	f, now := newTest(10)
	if ok, _ := f.Check(1); !ok {
		t.Fatal("sin historial debe permitir")
	}
	f.Fail(1)
	f.Fail(1)
	if ok, _ := f.Check(1); !ok {
		t.Fatal("bajo el umbral debe permitir")
	}
	f.Fail(1) // 3er fallo: 30 s
	if ok, w := f.Check(1); ok || w != 30*time.Second {
		t.Fatalf("esperaba bloqueo 30s, ok=%v w=%v", ok, w)
	}
	*now = now.Add(31 * time.Second)
	if ok, _ := f.Check(1); !ok {
		t.Fatal("tras la espera debe permitir")
	}
	f.Fail(1) // 4º: 60 s
	if _, w := f.Check(1); w != time.Minute {
		t.Fatalf("esperaba 60s, got %v", w)
	}
	*now = now.Add(61 * time.Second)
	f.Fail(1) // 5º: 120 s (tope)
	f.Fail(1) // 6º: seguiría creciendo pero se limita a MaxWait
	if _, w := f.Check(1); w != 2*time.Minute {
		t.Fatalf("esperaba tope 2m, got %v", w)
	}
}

func TestResetYVentanaVencida(t *testing.T) {
	f, now := newTest(10)
	for i := 0; i < 3; i++ {
		f.Fail(1)
	}
	f.Reset(1)
	if ok, _ := f.Check(1); !ok {
		t.Fatal("Reset debe limpiar el bloqueo")
	}
	f.Fail(2)
	f.Fail(2)
	*now = now.Add(16 * time.Minute) // ventana vencida: el conteo reinicia
	f.Fail(2)
	f.Fail(2)
	if ok, _ := f.Check(2); !ok {
		t.Fatal("con la ventana vencida los fallos antiguos no cuentan")
	}
}

func TestMemoriaAcotada(t *testing.T) {
	f, now := newTest(2)
	f.Fail(1)
	f.Fail(2)
	f.Fail(3) // lleno y vigentes: no se registra
	if len(f.m) != 2 {
		t.Fatalf("esperaba 2 entradas, hay %d", len(f.m))
	}
	*now = now.Add(16 * time.Minute)
	f.Fail(3) // se purgan las vencidas
	if len(f.m) != 1 {
		t.Fatalf("esperaba 1 entrada tras purgar, hay %d", len(f.m))
	}
	// una entrada vencida pero aún bloqueada no se purga
	f2, now2 := newTest(1)
	f2.cfg.BaseWait = time.Hour
	f2.cfg.MaxWait = time.Hour
	for i := 0; i < 3; i++ {
		f2.Fail(1)
	}
	*now2 = now2.Add(16 * time.Minute)
	f2.Fail(2)
	if _, ok := f2.m[1]; !ok {
		t.Fatal("no debe purgarse una entrada aún bloqueada")
	}
}

func TestConcurrencia(t *testing.T) {
	f := NewFailures(Config{MaxFailures: 5, Window: time.Minute, BaseWait: time.Second, MaxWait: time.Minute, MaxEntries: 50})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int64) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				f.Check(i % 5)
				f.Fail(i % 5)
				if j%10 == 0 {
					f.Reset(i % 5)
				}
			}
		}(int64(i))
	}
	wg.Wait()
}

func TestRetryAfterSeconds(t *testing.T) {
	for d, want := range map[time.Duration]int{0: 1, -time.Second: 1, 1: 1, time.Second: 1, 1500 * time.Millisecond: 2, 30 * time.Second: 30} {
		if got := RetryAfterSeconds(d); got != want {
			t.Fatalf("%v: got %d, want %d", d, got, want)
		}
	}
}

func TestWaitSeLimitaCuandoElDobleSuperaElTope(t *testing.T) {
	f := NewFailures(Config{MaxFailures: 3, Window: time.Hour, BaseWait: 30 * time.Second, MaxWait: 100 * time.Second, MaxEntries: 10})
	// 3.º fallo: 30 s; 4.º: 60 s; 5.º: 120 s > tope, se limita a 100 s.
	if got := f.wait(5); got != 100*time.Second {
		t.Fatalf("esperaba tope de 100s, got %v", got)
	}
}
