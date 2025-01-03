package models

type ErrorResponse struct {
	Message string `json:"message" example:"Internal server error"`
}
