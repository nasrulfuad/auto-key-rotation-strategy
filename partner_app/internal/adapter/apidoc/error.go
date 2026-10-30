package apidoc

type ErrorDetail struct {
	Code    string  `json:"code" example:"TOKEN_SIGNING_KEY_NOT_FOUND"`
	Field   *string `json:"field,omitempty"`
	Message string  `json:"message" example:"authentication required"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}
