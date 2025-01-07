package models

type TagResponse struct {
	Tags  []string `json:"tags"`
	Total int      `json:"total"`
}
