package models

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause,omitempty"`
	Data    any    `json:"data,omitempty"`
}
