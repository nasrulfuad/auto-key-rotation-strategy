package signingkey_test

import (
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/activationnotification"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

var _ activationnotification.TokenSigner = (*signingkey.Signer)(nil)
