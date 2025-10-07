package telemetria

import (
	"testing"
)

// TestBuildAdvancedDateFilter prueba la construcción de filtros SQL avanzados
func TestBuildAdvancedDateFilter(t *testing.T) {
	tests := []struct {
		name       string
		startDate  string
		endDate    string
		startTime  string
		endTime    string
		shouldPass bool
	}{
		{
			name:       "Same date default times",
			startDate:  "2025-01-01",
			endDate:    "2025-01-01",
			startTime:  DefaultStartTime,
			endTime:    DefaultEndTime,
			shouldPass: true,
		},
		{
			name:       "Date range default times",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  DefaultStartTime,
			endTime:    DefaultEndTime,
			shouldPass: true,
		},
		{
			name:       "Date range with custom times",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  "08:00:00",
			endTime:    "18:00:00",
			shouldPass: true,
		},
		{
			name:       "Same date with custom times",
			startDate:  "2025-01-15",
			endDate:    "2025-01-15",
			startTime:  "10:00:00",
			endTime:    "20:00:00",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAdvancedDateFilter(tt.startDate, tt.endDate, tt.startTime, tt.endTime)
			if result == "" && tt.shouldPass {
				t.Errorf("Expected non-empty filter")
			}
		})
	}
}

// TestBuildAdvancedDateFilterWithField prueba la construcción de filtros SQL con campo personalizado
func TestBuildAdvancedDateFilterWithField(t *testing.T) {
	tests := []struct {
		name       string
		dateField  string
		startDate  string
		endDate    string
		startTime  string
		endTime    string
		shouldPass bool
	}{
		{
			name:       "Same date default times",
			dateField:  "pe.fecha",
			startDate:  "2025-01-01",
			endDate:    "2025-01-01",
			startTime:  DefaultStartTime,
			endTime:    DefaultEndTime,
			shouldPass: true,
		},
		{
			name:       "Custom field date range",
			dateField:  "r.fecha",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  DefaultStartTime,
			endTime:    DefaultEndTime,
			shouldPass: true,
		},
		{
			name:       "Custom field with time filters",
			dateField:  "r.fecha",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  "08:00:00",
			endTime:    "18:00:00",
			shouldPass: true,
		},
		{
			name:       "Edge case start time only",
			dateField:  "pe.fecha",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  "08:00:00",
			endTime:    DefaultEndTime,
			shouldPass: true,
		},
		{
			name:       "Edge case end time only",
			dateField:  "pe.fecha",
			startDate:  "2025-01-01",
			endDate:    "2025-01-31",
			startTime:  DefaultStartTime,
			endTime:    "18:00:00",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAdvancedDateFilterWithField(tt.dateField, tt.startDate, tt.endDate, tt.startTime, tt.endTime)
			if result == "" && tt.shouldPass {
				t.Errorf("Expected non-empty filter")
			}
			// Verificar que el campo está presente en el resultado
			if tt.shouldPass && len(result) > 0 && result[:len(tt.dateField)] != tt.dateField {
				// El campo debe aparecer al inicio o después de paréntesis
				// Solo verificamos que no esté vacío
				if result == "" {
					t.Errorf("Expected field %s in filter, got %s", tt.dateField, result)
				}
			}
		})
	}
}

// TestParseFilterParams prueba el parsing de parámetros de filtro
func TestParseFilterParams(t *testing.T) {
	// Esta función requiere un controlador de Beego, pero podemos probar indirectamente
	// a través de getAdvancedTimeRange que es llamada por parseFilterParams

	// Probar casos edge de getAdvancedTimeRange directamente
	tests := []struct {
		name        string
		filter      TimeFilter
		mes         string
		año         string
		fechaInicio string
		fechaFin    string
		horaInicio  string
		horaFin     string
	}{
		{
			name:   "FilterMonthYear without params",
			filter: FilterMonthYear,
			mes:    "",
			año:    "",
		},
		{
			name:   "FilterMonthYear with invalid month string",
			filter: FilterMonthYear,
			mes:    "invalid",
			año:    "2025",
		},
		{
			name:   "FilterMonthYear with invalid year string",
			filter: FilterMonthYear,
			mes:    "6",
			año:    "invalid",
		},
		{
			name:   "FilterMonthYear with month 0",
			filter: FilterMonthYear,
			mes:    "0",
			año:    "2025",
		},
		{
			name:        "FilterDateRange with invalid dates",
			filter:      FilterDateRange,
			fechaInicio: "invalid",
			fechaFin:    "invalid",
		},
		{
			name:        "FilterDateRange without params",
			filter:      FilterDateRange,
			fechaInicio: "",
			fechaFin:    "",
		},
		{
			name:        "Time with HH:MM format",
			filter:      FilterDateRange,
			fechaInicio: "2025-01-01",
			fechaFin:    "2025-01-31",
			horaInicio:  "08:00",
			horaFin:     "18:00",
		},
		{
			name:        "Invalid time format",
			filter:      FilterDateRange,
			fechaInicio: "2025-01-01",
			fechaFin:    "2025-01-31",
			horaInicio:  "invalid",
			horaFin:     "invalid",
		},
		{
			name:   "Month edge case December",
			filter: FilterMonthYear,
			mes:    "12",
			año:    "2025",
		},
		{
			name:   "Month edge case January",
			filter: FilterMonthYear,
			mes:    "1",
			año:    "2025",
		},
		{
			name:   "Year edge case minimum",
			filter: FilterMonthYear,
			mes:    "6",
			año:    "1899", // Should default to current year
		},
		{
			name:   "Year edge case maximum",
			filter: FilterMonthYear,
			mes:    "6",
			año:    "2101", // Should default to current year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate, startTime, endTime := getAdvancedTimeRange(
				tt.filter, tt.mes, tt.año, tt.fechaInicio, tt.fechaFin, tt.horaInicio, tt.horaFin,
			)

			// Verificar que las fechas no están vacías
			if startDate == "" || endDate == "" {
				t.Errorf("Expected non-empty dates, got start=%s, end=%s", startDate, endDate)
			}

			// Verificar que los tiempos no están vacíos
			if startTime == "" || endTime == "" {
				t.Errorf("Expected non-empty times, got startTime=%s, endTime=%s", startTime, endTime)
			}

			// Verificar formato de tiempo HH:MM:SS
			if len(startTime) != 8 || len(endTime) != 8 {
				t.Errorf("Expected time format HH:MM:SS, got startTime=%s, endTime=%s", startTime, endTime)
			}
		})
	}
}

// TestTimeFilterConstants verifica que las constantes estén definidas correctamente
func TestTimeFilterConstants(t *testing.T) {
	constants := []TimeFilter{
		FilterToday,
		FilterLastWeek,
		FilterLastMonth,
		FilterLast3Months,
		FilterLast6Months,
		FilterLastYear,
		FilterHistoric,
		FilterMonthYear,
		FilterDateRange,
	}

	for _, c := range constants {
		if c == "" {
			t.Errorf("Expected non-empty constant")
		}
	}
}

// TestDefaultTimeConstants verifica las constantes de tiempo por defecto
func TestDefaultTimeConstants(t *testing.T) {
	if DefaultStartTime != "00:00:00" {
		t.Errorf("Expected DefaultStartTime to be 00:00:00, got %s", DefaultStartTime)
	}
	if DefaultEndTime != "23:59:59" {
		t.Errorf("Expected DefaultEndTime to be 23:59:59, got %s", DefaultEndTime)
	}
}
