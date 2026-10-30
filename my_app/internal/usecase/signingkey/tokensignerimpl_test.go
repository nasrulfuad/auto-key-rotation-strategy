package signingkey_test

import (
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/signingkey"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/subscription"
)

var _ subscription.TokenSigner = (*signingkey.Signer)(nil)
