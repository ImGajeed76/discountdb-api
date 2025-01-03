package models

type CategoriesResponse struct {
	Total int      `json:"total" example:"2"`
	Data  []string `json:"data" example:[\"Electronics\",\"Clothing\"]`
}
