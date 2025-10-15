package models

import (
	"testing"
	"time"
)

func TestParseDateToNoonUTC(t *testing.T) {
	got, err := ParseDateToNoonUTC("2025-10-14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", got.Location())
	}
	if got.Year() != 2025 || got.Month() != time.October || got.Day() != 14 {
		t.Fatalf("expected 2025-10-14, got %v", got)
	}
	if got.Hour() != 12 || got.Minute() != 0 || got.Second() != 0 {
		t.Fatalf("expected 12:00:00, got %02d:%02d:%02d", got.Hour(), got.Minute(), got.Second())
	}
}

func TestParseDateToNoonUTCInvalid(t *testing.T) {
	if _, err := ParseDateToNoonUTC("2025-13-01"); err == nil {
		t.Fatalf("expected error for invalid date")
	}
}

func TestParseTimeToUTCWithSeconds(t *testing.T) {
	got, err := ParseTimeToUTC("08:30:15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", got.Location())
	}
	if got.Year() != 1 || got.Month() != time.January || got.Day() != 1 {
		t.Fatalf("expected base date 0001-01-01, got %v", got)
	}
	if got.Hour() != 8 || got.Minute() != 30 || got.Second() != 15 {
		t.Fatalf("expected 08:30:15, got %02d:%02d:%02d", got.Hour(), got.Minute(), got.Second())
	}
}

func TestParseTimeToUTCMinutesOnly(t *testing.T) {
	got, err := ParseTimeToUTC("08:30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Hour() != 8 || got.Minute() != 30 || got.Second() != 0 {
		t.Fatalf("expected 08:30:00, got %02d:%02d:%02d", got.Hour(), got.Minute(), got.Second())
	}
}

func TestFormatTimeWithLMTAdjustment(t *testing.T) {

	t0 := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	s := FormatTimeWithLMT(t0)
	if s != "09:52:32" {
		t.Fatalf("expected 09:52:32 for LMT adjustment, got %s", s)
	}

	tOther := time.Date(2000, 1, 1, 10, 5, 7, 0, time.UTC)
	s2 := FormatTimeWithLMT(tOther)
	if s2 != "10:05:07" {
		t.Fatalf("expected 10:05:07 without adjustment, got %s", s2)
	}
}
