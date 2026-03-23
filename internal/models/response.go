package models

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    int    `json:"errorCode"`
	ErrorDetails string `json:"errorDetails,omitempty"`
}
