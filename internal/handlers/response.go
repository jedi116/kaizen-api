// internal/handlers/response.go
package handlers

type ErrorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Success"`
}
