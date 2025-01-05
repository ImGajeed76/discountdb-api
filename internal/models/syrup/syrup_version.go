package syrup

type VersionInfo struct {
	Version  string `json:"version" example:"1.0.0"`
	Provider string `json:"provider" example:"DiscountDB"`
}
