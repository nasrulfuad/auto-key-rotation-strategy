package signingkey_test

import (
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/signingkey"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/subscription"
)

var _ subscription.TokenSigner = (*signingkey.Signer)(nil)
