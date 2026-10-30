package domain

import (
	"slices"
	"time"
)

type TokenAudience string

type TokenClaims struct {
	ID        string
	Issuer    string
	Subject   string
	Provider  string
	Audience  []string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type TokenRequest struct {
	Audience TokenAudience
	Subject  string
	Provider string
}

type SignedToken struct {
	Value     string
	KeyID     string
	Claims    TokenClaims
	ExpiresAt time.Time
}

type VerifiedToken struct {
	Claims TokenClaims
	KeyID  string
}

type TokenExpectation struct {
	Issuer   string
	Audience TokenAudience
	Leeway   time.Duration
}

func (e TokenExpectation) Check(claims TokenClaims, now time.Time) error {
	if claims.Issuer != e.Issuer {
		return ErrTokenIssuerUnknown
	}
	if !slices.Contains(claims.Audience, string(e.Audience)) {
		return TokenClaimInvalid("audience_mismatch")
	}
	if claims.ExpiresAt.IsZero() || !now.Before(claims.ExpiresAt.Add(e.Leeway)) {
		return TokenClaimInvalid("token_expired")
	}
	if claims.IssuedAt.IsZero() || claims.IssuedAt.After(now.Add(e.Leeway)) {
		return TokenClaimInvalid("token_issued_in_future")
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt) > MaximumTokenTTL {
		return TokenClaimInvalid("token_lifetime_too_long")
	}
	return nil
}
