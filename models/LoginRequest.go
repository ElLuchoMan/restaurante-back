package models

type LoginRequest struct {
	Documento int    `json:"documento"`
	Password  string `json:"password"`
}
