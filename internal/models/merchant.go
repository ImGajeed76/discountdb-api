package models

type Merchant struct {
	Name string `json:"merchant_name" example:"merchant1"`
	URL  string `json:"merchant_url" example:"https://merchant1.com"`
}

type MerchantResponse struct {
	Total int        `json:"total" example:"2"`
	Data  []Merchant `json:"data" example:[{"merchant_name":"merchant1","merchant_url":"https://merchant1.com"},{"merchant_name":"merchant2","merchant_url":"https://merchant2.com"}]`
}
