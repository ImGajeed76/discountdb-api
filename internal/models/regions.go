package models

type RegionResponse struct {
	Regions []string `json:"regions"`
	Total   int      `json:"total"`
}
