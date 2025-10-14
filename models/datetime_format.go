package models

import (
	"fmt"
	"time"
)

func FormatDateUTC(t time.Time) string {
	utc := t.UTC()
	return fmt.Sprintf("%02d-%02d-%04d", utc.Day(), int(utc.Month()), utc.Year())
}

func FormatTimeWithLMT(t time.Time) string {
	adj := t
	if adj.Year() < 1900 {
		adj = adj.Add(9*time.Hour + 52*time.Minute + 32*time.Second)
	}
	return fmt.Sprintf("%02d:%02d:%02d", adj.Hour(), adj.Minute(), adj.Second())
}

func FormatTimestampBogota(t time.Time) string {
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}
	return t.In(loc).Format("02-01-2006 15:04:05")
}

// ParseDateToNoonUTC parsea una fecha YYYY-MM-DD y devuelve el instante a las 12:00:00 UTC del mismo día
func ParseDateToNoonUTC(s string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	// Usamos componentes de calendario y fijamos mediodía en UTC para evitar corrimientos
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC), nil
}

// ParseTimeToUTC parsea una hora HH:MM[:SS] y la normaliza a UTC en la fecha base 0001-01-01
func ParseTimeToUTC(s string) (time.Time, error) {
	// aceptar HH:MM:SS o HH:MM
	var (
		t   time.Time
		err error
	)
	if len(s) == len("15:04:05") {
		t, err = time.Parse("15:04:05", s)
	} else {
		// fallback a HH:MM
		t, err = time.Parse("15:04", s)
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC), nil
}
