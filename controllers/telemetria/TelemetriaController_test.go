package telemetria

import (
	"net/http"
	"net/http/httptest"
	"testing"

	loginc "restaurante/controllers/login"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func TestTelemetriaController_GetDashboard_Unauthorized(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/dashboard", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetDashboard()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetSales_Unauthorized(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/sales", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetSales()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetProducts_Unauthorized(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/products", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetProducts()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetUsers_Unauthorized(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/users", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetUsers()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetTimeAnalysis_Unauthorized(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/time-analysis", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetTimeAnalysis()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetJWTSecret(t *testing.T) {

	web.BConfig.RunMode = "dev"

	secret := loginc.GetJWTSecret()

	if len(secret) == 0 {
		t.Error("Expected non-empty JWT secret")
	}
}

func TestValidateAdminRole_InvalidToken(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/dashboard", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	_, valid := controller.validateAdminRole()

	if valid {
		t.Error("Expected validation to fail with invalid token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetTimeRange(t *testing.T) {
	tests := []struct {
		name   string
		filter TimeFilter
	}{
		{"FilterToday", FilterToday},
		{"FilterLastWeek", FilterLastWeek},
		{"FilterLastMonth", FilterLastMonth},
		{"FilterLast3Months", FilterLast3Months},
		{"FilterLast6Months", FilterLast6Months},
		{"FilterLastYear", FilterLastYear},
		{"FilterHistoric", FilterHistoric},
		{"FilterDefault", TimeFilter("invalid")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate := getTimeRange(tt.filter)

			if startDate == "" || endDate == "" {
				t.Errorf("Expected non-empty dates, got start=%s, end=%s", startDate, endDate)
			}

			if len(startDate) != 10 || len(endDate) != 10 {
				t.Errorf("Expected date format YYYY-MM-DD, got start=%s, end=%s", startDate, endDate)
			}
		})
	}
}

func TestBuildDateFilter(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		expected  string
	}{
		{
			name:      "Same date",
			startDate: "2025-01-01",
			endDate:   "2025-01-01",
			expected:  "pe.fecha = '2025-01-01'",
		},
		{
			name:      "Date range",
			startDate: "2025-01-01",
			endDate:   "2025-01-31",
			expected:  "pe.fecha >= '2025-01-01' AND pe.fecha <= '2025-01-31'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDateFilter(tt.startDate, tt.endDate)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetAdvancedTimeRange(t *testing.T) {
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
			name:   "FilterMonthYear with valid month and year",
			filter: FilterMonthYear,
			mes:    "6",
			año:    "2025",
		},
		{
			name:   "FilterMonthYear with invalid month",
			filter: FilterMonthYear,
			mes:    "13",
			año:    "2025",
		},
		{
			name:        "FilterDateRange with valid dates",
			filter:      FilterDateRange,
			fechaInicio: "2025-01-01",
			fechaFin:    "2025-01-31",
		},
		{
			name:        "FilterDateRange with times",
			filter:      FilterDateRange,
			fechaInicio: "2025-01-01",
			fechaFin:    "2025-01-31",
			horaInicio:  "08:00:00",
			horaFin:     "18:00:00",
		},
		{
			name:   "Other filter",
			filter: FilterToday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, endDate, startTime, endTime := getAdvancedTimeRange(
				tt.filter, tt.mes, tt.año, tt.fechaInicio, tt.fechaFin, tt.horaInicio, tt.horaFin,
			)

			if startDate == "" || endDate == "" {
				t.Errorf("Expected non-empty dates, got start=%s, end=%s", startDate, endDate)
			}

			if startTime == "" || endTime == "" {
				t.Errorf("Expected non-empty times, got startTime=%s, endTime=%s", startTime, endTime)
			}
		})
	}
}
