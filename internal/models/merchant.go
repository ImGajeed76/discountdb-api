package models

type Merchant struct {
	Name    string   `json:"merchant_name" example:"merchant1"`
	Domains []string `json:"merchant_url" example:["https://merchant1.com"]`
}

type MerchantResponse struct {
	Total int        `json:"total" example:"2"`
	Data  []Merchant `json:"data"`
}
