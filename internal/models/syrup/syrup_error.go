package syrup

type ErrorResponse struct {
	Error   string `json:"error" example:"Internal Server Error"`
	Message string `json:"message" example:"Something went wrong"`
}
