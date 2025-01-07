package models

type CategoriesResponse struct {
	Total      int      `json:"total" example:"2"`
	Categories []string `json:"data" example:[\"Electronics\",\"Clothing\"]`
}
