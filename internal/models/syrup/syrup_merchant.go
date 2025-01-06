package syrup

type Merchant struct {
	MerchantName string   `json:"merchant_name"`
	Domains      []string `json:"domains"`
}

type MerchantList struct {
	Merchants []Merchant `json:"merchants"`
	Total     int        `json:"total"`
}
