package entitlement

import "github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"

func ensureProvider(token domain.VerifiedToken, peerProvider string) error {
	if token.Claims.Provider == "" || token.Claims.Provider != peerProvider {
		return domain.TokenClaimInvalid("provider_not_allowed")
	}
	return nil
}
