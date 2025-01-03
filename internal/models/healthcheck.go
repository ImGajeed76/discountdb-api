package models

type HealthCheckResponse struct {
	Status  string `json:"status" example:"ok"`
	Version string `json:"version" example:"1.0"`
}
