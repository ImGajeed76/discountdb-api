package syrup

type Coupon struct {
	ID           string  `json:"id" example:"123"`
	Title        string  `json:"title" example:"Discount"`
	Description  string  `json:"description" example:"Get 10% off"`
	Code         string  `json:"code" example:"DISCOUNT10"`
	Score        float64 `json:"score" example:"5"`
	MerchantName string  `json:"merchant_name" example:"Amazon"`
}

type CouponList struct {
	Coupons []Coupon `json:"coupons"`
	Total   int      `json:"total"`
}
