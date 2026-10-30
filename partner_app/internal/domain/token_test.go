package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

func TestTokenExpectationCheck(t *testing.T) {
	expectation := domain.TokenExpectation{Issuer: "qa.peer.local", Audience: "get.entitlement", Leeway: 30 * time.Second}
	valid := domain.TokenClaims{
		Issuer:    "qa.peer.local",
		Audience:  []string{"get.entitlement"},
		IssuedAt:  baseTime,
		ExpiresAt: baseTime.Add(5 * time.Minute),
	}
	cases := []struct {
		name   string
		mutate func(*domain.TokenClaims)
		now    time.Time
		want   error
	}{
		{"valid", func(*domain.TokenClaims) {}, baseTime.Add(time.Minute), nil},
		{"wrong issuer", func(c *domain.TokenClaims) { c.Issuer = "prod.peer.local" }, baseTime, domain.ErrTokenIssuerUnknown},
		{"wrong audience", func(c *domain.TokenClaims) { c.Audience = []string{"set.entitlement"} }, baseTime, domain.TokenClaimInvalid("audience_mismatch")},
		{"expired beyond leeway", func(*domain.TokenClaims) {}, baseTime.Add(5*time.Minute + 30*time.Second), domain.TokenClaimInvalid("token_expired")},
		{"expired within leeway", func(*domain.TokenClaims) {}, baseTime.Add(5*time.Minute + 29*time.Second), nil},
		{"issued in the future", func(*domain.TokenClaims) {}, baseTime.Add(-31 * time.Second), domain.TokenClaimInvalid("token_issued_in_future")},
		{"missing expiry", func(c *domain.TokenClaims) { c.ExpiresAt = time.Time{} }, baseTime, domain.TokenClaimInvalid("token_expired")},
		{"lifetime above 24h", func(c *domain.TokenClaims) { c.ExpiresAt = c.IssuedAt.Add(25 * time.Hour) }, baseTime, domain.TokenClaimInvalid("token_lifetime_too_long")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := valid
			tc.mutate(&claims)
			err := expectation.Check(claims, tc.now)
			if tc.want == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.want)
		})
	}
}
