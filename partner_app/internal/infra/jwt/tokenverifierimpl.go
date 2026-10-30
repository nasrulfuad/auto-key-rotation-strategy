package jwt

import (
	"context"
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
)

var _ PublicKeySource = (*jwks.RemoteKeySet)(nil)

type Verifier struct {
	publicKeySource PublicKeySource
	algorithm       domain.SigningAlgorithm
	logger          logging.Logger
}

var _ TokenVerifier = (*Verifier)(nil)

func NewVerifier(publicKeySource PublicKeySource, algorithm domain.SigningAlgorithm, logger logging.Logger) *Verifier {
	if logger == nil {
		logger = logging.NopLogger{}
	}
	return &Verifier{publicKeySource: publicKeySource, algorithm: algorithm, logger: logger}
}

func (v *Verifier) Verify(ctx context.Context, raw string, expectation domain.TokenExpectation, now time.Time) (domain.VerifiedToken, error) {
	parser := gojwt.NewParser(
		gojwt.WithValidMethods([]string{string(v.algorithm)}),
		gojwt.WithoutClaimsValidation(),
	)
	var claims tokenClaims
	var keyID string
	keyIDMissing := false
	_, err := parser.ParseWithClaims(raw, &claims, func(token *gojwt.Token) (any, error) {
		if claims.Issuer != expectation.Issuer {
			return nil, domain.ErrTokenIssuerUnknown
		}
		keyID, _ = token.Header["kid"].(string)
		publicKey, err := v.publicKeySource.FindPublicKey(ctx, keyID)
		if err == nil {
			return publicKey, nil
		}
		keyIDMissing = true
		v.logger.Warn(ctx, "token kid not found in the peer key set; trying every cached key", "kid", keyID, "error", err)
		verificationKeys := gojwt.VerificationKeySet{}
		for _, candidate := range v.publicKeySource.ListAllPublicKeys() {
			verificationKeys.Keys = append(verificationKeys.Keys, candidate)
		}
		if len(verificationKeys.Keys) == 0 {
			return nil, domain.ErrTokenSigningKeyNotFound
		}
		return verificationKeys, nil
	})
	if err != nil {
		return domain.VerifiedToken{}, toVerificationError(err, keyIDMissing)
	}
	verified := domain.VerifiedToken{Claims: toDomainClaims(claims), KeyID: keyID}
	if err := expectation.Check(verified.Claims, now); err != nil {
		return domain.VerifiedToken{}, err
	}
	return verified, nil
}

func toVerificationError(err error, keyIDMissing bool) error {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, gojwt.ErrTokenSignatureInvalid) && keyIDMissing {
		return domain.ErrTokenSigningKeyNotFound
	}
	if errors.Is(err, gojwt.ErrTokenSignatureInvalid) {
		return domain.TokenInvalid("signature_invalid")
	}
	if errors.Is(err, gojwt.ErrTokenUnverifiable) {
		return domain.TokenInvalid("token_unverifiable")
	}
	return domain.TokenInvalid("token_malformed")
}

func toDomainClaims(claims tokenClaims) domain.TokenClaims {
	result := domain.TokenClaims{
		ID:       claims.ID,
		Issuer:   claims.Issuer,
		Subject:  claims.Subject,
		Provider: claims.Provider,
		Audience: claims.Audience,
	}
	if claims.IssuedAt != nil {
		result.IssuedAt = claims.IssuedAt.UTC()
	}
	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.UTC()
	}
	return result
}
