package models

import "time"

type DiscountType string

const (
	PercentageOff DiscountType = "PERCENTAGE_OFF"
	FixedAmount   DiscountType = "FIXED_AMOUNT"
	BOGO          DiscountType = "BOGO"
	FreeShipping  DiscountType = "FREE_SHIPPING"
)

type Coupon struct {
	// Required Information
	ID            int64        `json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	DiscountValue float64      `json:"discount_value"`
	DiscountType  DiscountType `json:"discount_type"`
	MerchantName  string       `json:"merchant_name"`
	MerchantURL   string       `json:"merchant_url"`

	// Optional Validity Information
	StartDate             *time.Time `json:"start_date,omitempty"`
	EndDate               *time.Time `json:"end_date,omitempty"`
	TermsConditions       string     `json:"terms_conditions,omitempty"`
	MinimumPurchaseAmount float64    `json:"minimum_purchase_amount,omitempty"`
	MaximumDiscountAmount float64    `json:"maximum_discount_amount,omitempty"`

	// Voting Information
	UpVotes   TimestampArray `json:"up_votes"`
	DownVotes TimestampArray `json:"down_votes"`

	// Metadata
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Regions    []string `json:"regions,omitempty"`    // countries/regions where valid
	StoreType  string   `json:"store_type,omitempty"` // "online", "in_store", "both"

	// Score calculated by db
	MaterializedScore float64    `json:"score"`
	LastScoreUpdate   *time.Time `json:"-"` // not exposed to API
}

type CouponsSearchResponse struct {
	Data   []Coupon `json:"data" example:[{"id":1,"title":"Discount","description":"Get 10% off","score":5,"created_at":"2021-01-01T00:00:00Z"}]`
	Total  int      `json:"total" example:"100"`
	Limit  int      `json:"limit" example:"10"`
	Offset int      `json:"offset" example:"0"`
}

type CouponsSearchParams struct {
	SearchString string `json:"search_string" example:"discount"`
	SortBy       string `json:"sort_by" example:"newest" enums:"newest,oldest,high_score,low_score"`
	Limit        int    `json:"limit" example:"10" minimum:"1"`
	Offset       int    `json:"offset" example:"0" minimum:"0"`
}

type CouponCreateRequest struct {
	// Required Information
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	DiscountValue float64      `json:"discount_value"`
	DiscountType  DiscountType `json:"discount_type"`
	MerchantName  string       `json:"merchant_name"`
	MerchantURL   string       `json:"merchant_url"`

	// Optional Validity Information
	StartDate             *time.Time `json:"start_date,omitempty"`
	EndDate               *time.Time `json:"end_date,omitempty"`
	TermsConditions       string     `json:"terms_conditions,omitempty"`
	MinimumPurchaseAmount float64    `json:"minimum_purchase_amount,omitempty"`
	MaximumDiscountAmount float64    `json:"maximum_discount_amount,omitempty"`

	// Metadata
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Regions    []string `json:"regions,omitempty"`    // countries/regions where valid
	StoreType  string   `json:"store_type,omitempty"` // "online", "in_store", "both"
}

type CouponCreateResponse struct {
	ID                int64     `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	MaterializedScore float64   `json:"score"`
}
