package signingkey_test

import (
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/activationnotification"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/signingkey"
)

var _ activationnotification.TokenSigner = (*signingkey.Signer)(nil)
