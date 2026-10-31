package rsakey_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/rsakey"
)

func TestThumbprintMatchesRFC7638Example(t *testing.T) {
	modulus, err := rsakey.DecodeBase64URL("0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw")
	require.NoError(t, err)
	exponent, err := rsakey.DecodeBase64URL("AQAB")
	require.NoError(t, err)

	thumbprint, err := rsakey.Thumbprint(domain.PublicKey{Modulus: modulus, Exponent: exponent})

	require.NoError(t, err)
	assert.Equal(t, "NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs", thumbprint)
}

func TestGeneratorProducesDistinctKeyIDsAndParsablePrivateKeys(t *testing.T) {
	generator, err := rsakey.NewGenerator(rsakey.WithBits(2048))
	require.NoError(t, err)

	first, err := generator.Generate(context.Background(), domain.SigningAlgorithmRS256)
	require.NoError(t, err)
	second, err := generator.Generate(context.Background(), domain.SigningAlgorithmRS256)
	require.NoError(t, err)

	assert.NotEqual(t, first.KeyID, second.KeyID)
	privateKey, err := rsakey.ParsePrivateKey(first.PrivateKey)
	require.NoError(t, err)
	assert.Equal(t, first.PublicKey, rsakey.ToDomainPublicKey(&privateKey.PublicKey))
	thumbprint, err := rsakey.Thumbprint(first.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, thumbprint, first.KeyID)
}

func TestNewGeneratorRejectsWeakKeys(t *testing.T) {
	_, err := rsakey.NewGenerator(rsakey.WithBits(1024))
	assert.Error(t, err)
}
