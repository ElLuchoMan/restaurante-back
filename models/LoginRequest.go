package models

type LoginRequest struct {
	Documento int64  `json:"documento"`
	Password  string `json:"password"`
}
