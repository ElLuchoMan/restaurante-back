package models

import (
	"encoding/json"
	"testing"
)

// Tests para structs de Request/Response simples

func TestLoginRequest_Marshal(t *testing.T) {
	req := LoginRequest{
		Documento: 123456789,
		Password:  "test",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Error marshaling LoginRequest: %v", err)
	}

	var decoded LoginRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Error unmarshaling LoginRequest: %v", err)
	}

	if decoded.Documento != req.Documento {
		t.Errorf("Expected documento %d, got %d", req.Documento, decoded.Documento)
	}
}

func TestApiResponse_Success(t *testing.T) {
	resp := ApiResponse{
		Code:    200,
		Message: "Success",
		Data:    map[string]string{"key": "value"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Error marshaling ApiResponse: %v", err)
	}

	var decoded ApiResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Error unmarshaling ApiResponse: %v", err)
	}

	if decoded.Code != 200 {
		t.Errorf("Expected code 200, got %d", decoded.Code)
	}
}

func TestAuthResponse_Tokens(t *testing.T) {
	resp := map[string]string{
		"token":         "access_token_123",
		"access_token":  "access_token_123",
		"refresh_token": "refresh_token_456",
		"token_type":    "Bearer",
		"expires_in":    "3600",
		"nombre":        "Test User",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Error marshaling AuthResponse: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestPaginatedResponse_WithData(t *testing.T) {
	resp := PaginatedResponse{
		Data:       []string{"item1", "item2"},
		Total:      100,
		Page:       1,
		PageSize:   10,
		TotalPages: 10,
	}

	if resp.TotalPages != 10 {
		t.Errorf("Expected 10 total pages, got %d", resp.TotalPages)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Error marshaling PaginatedResponse: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestNotificacionRequest_Marshal(t *testing.T) {
	req := NotificacionRequest{
		Titulo:  "Test Notification",
		Mensaje: "Test message",
		Datos: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Error marshaling NotificacionRequest: %v", err)
	}

	var decoded NotificacionRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Error unmarshaling NotificacionRequest: %v", err)
	}

	if decoded.Titulo != req.Titulo {
		t.Errorf("Expected titulo %s, got %s", req.Titulo, decoded.Titulo)
	}
}

func TestNotificacionResponse_Success(t *testing.T) {
	resp := NotificacionResponse{
		IDEnvio:     123,
		Exitosos:    5,
		Fallidos:    0,
		Mensajes:    []string{"Success"},
		TokensError: []string{},
	}

	if resp.Exitosos != 5 {
		t.Errorf("Expected 5 exitosos, got %d", resp.Exitosos)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Error marshaling NotificacionResponse: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

func TestNotificacionResponse_WithErrors(t *testing.T) {
	resp := NotificacionResponse{
		IDEnvio:     456,
		Exitosos:    3,
		Fallidos:    2,
		Mensajes:    []string{"Error 1", "Error 2"},
		TokensError: []string{"token1", "token2"},
	}

	if len(resp.TokensError) != 2 {
		t.Errorf("Expected 2 error tokens, got %d", len(resp.TokensError))
	}
}

func TestApiResponse_WithError(t *testing.T) {
	resp := ApiResponse{
		Code:    500,
		Message: "Internal Server Error",
		Cause:   "Database connection failed",
	}

	if resp.Cause != "Database connection failed" {
		t.Errorf("Expected specific cause, got %s", resp.Cause)
	}
}

func TestPaginatedResponse_EmptyData(t *testing.T) {
	resp := PaginatedResponse{
		Data:       []string{},
		Total:      0,
		Page:       1,
		PageSize:   10,
		TotalPages: 0,
	}

	if resp.Total != 0 {
		t.Errorf("Expected 0 total, got %d", resp.Total)
	}
}

func TestAuthResponse_EmptyFields(t *testing.T) {
	resp := map[string]string{}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Error marshaling empty map: %v", err)
	}

	if len(data) < 2 { // At least {}
		t.Error("Expected valid JSON for empty struct")
	}
}

func TestNotificacionRequest_EmptyDatos(t *testing.T) {
	req := NotificacionRequest{
		Titulo:  "Test",
		Mensaje: "Message",
		Datos:   nil,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Error marshaling NotificacionRequest with nil Datos: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}
