package jwt

import (
	"crypto/rsa"
	"fmt"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

func Sign(claims domain.TokenClaims, keyID string, algorithm domain.SigningAlgorithm, privateKey *rsa.PrivateKey) (string, error) {
	method := gojwt.GetSigningMethod(string(algorithm))
	if method == nil || !algorithm.Valid() {
		return "", fmt.Errorf("unsupported signing algorithm %q", algorithm)
	}
	token := gojwt.NewWithClaims(method, tokenClaims{
		Provider: claims.Provider,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        claims.ID,
			Issuer:    claims.Issuer,
			Subject:   claims.Subject,
			Audience:  claims.Audience,
			IssuedAt:  gojwt.NewNumericDate(claims.IssuedAt),
			ExpiresAt: gojwt.NewNumericDate(claims.ExpiresAt),
		},
	})
	token.Header["kid"] = keyID
	return token.SignedString(privateKey)
}
