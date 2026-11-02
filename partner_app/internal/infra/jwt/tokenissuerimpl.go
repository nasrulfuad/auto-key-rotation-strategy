package jwt

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/rsakey"
)

type Issuer struct {
	signingKeyReader repository.SigningKeyReader
}

var _ TokenIssuer = (*Issuer)(nil)

func NewIssuer(signingKeyReader repository.SigningKeyReader, opts ...Option) *Issuer {
	return &Issuer{signingKeyReader: signingKeyReader}
}

func (i *Issuer) Issue(ctx context.Context, claims domain.TokenClaims, key domain.SigningKey) (string, error) {
	secret, err := i.signingKeyReader.FindPrivateKey(ctx, key.KeyID)
	if err != nil {
		return "", err
	}
	privateKey, err := rsakey.ParsePrivateKey(secret)
	if err != nil {
		return "", err
	}
	return Sign(claims, key.KeyID, key.Algorithm, privateKey)
}
