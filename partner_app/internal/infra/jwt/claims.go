package jwt

import (
	gojwt "github.com/golang-jwt/jwt/v5"
)

type tokenClaims struct {
	Provider string `json:"provider,omitempty"`
	gojwt.RegisteredClaims
}
