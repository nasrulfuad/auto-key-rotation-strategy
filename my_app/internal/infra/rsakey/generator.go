package rsakey

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

const MinimumBits = 2048

type Generator struct {
	bits int
}

var _ port.KeyPairGenerator = Generator{}

func NewGenerator(bits int) (Generator, error) {
	if bits < MinimumBits {
		return Generator{}, fmt.Errorf("RSA key size %d is below the minimum of %d bits", bits, MinimumBits)
	}
	return Generator{bits: bits}, nil
}

func (g Generator) Generate(_ context.Context, algorithm domain.SigningAlgorithm) (domain.KeyPair, error) {
	if !algorithm.Valid() {
		return domain.KeyPair{}, fmt.Errorf("unsupported signing algorithm %q", algorithm)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, g.bits)
	if err != nil {
		return domain.KeyPair{}, err
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return domain.KeyPair{}, err
	}
	publicKey := ToDomainPublicKey(&privateKey.PublicKey)
	keyID, err := Thumbprint(publicKey)
	if err != nil {
		return domain.KeyPair{}, err
	}
	return domain.KeyPair{
		KeyID:      keyID,
		Algorithm:  algorithm,
		PublicKey:  publicKey,
		PrivateKey: domain.NewSecret(encoded),
	}, nil
}

func ToDomainPublicKey(publicKey *rsa.PublicKey) domain.PublicKey {
	return domain.PublicKey{
		Modulus:  publicKey.N.Bytes(),
		Exponent: big.NewInt(int64(publicKey.E)).Bytes(),
	}
}

func ToRSAPublicKey(publicKey domain.PublicKey) (*rsa.PublicKey, error) {
	exponent := new(big.Int).SetBytes(publicKey.Exponent)
	if len(publicKey.Modulus) == 0 || !exponent.IsInt64() || exponent.Int64() <= 1 || exponent.Int64() > 1<<31-1 {
		return nil, fmt.Errorf("invalid RSA public key")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(publicKey.Modulus), E: int(exponent.Int64())}, nil
}

func ParsePrivateKey(secret domain.Secret) (*rsa.PrivateKey, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(secret.Reveal())
	if err != nil {
		return nil, err
	}
	privateKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is %T, want RSA", parsed)
	}
	return privateKey, nil
}

func EncodeBase64URL(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func DecodeBase64URL(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}

type thumbprintInput struct {
	Exponent string `json:"e"`
	KeyType  string `json:"kty"`
	Modulus  string `json:"n"`
}

func Thumbprint(publicKey domain.PublicKey) (string, error) {
	canonical, err := json.Marshal(thumbprintInput{
		Exponent: EncodeBase64URL(publicKey.Exponent),
		KeyType:  "RSA",
		Modulus:  EncodeBase64URL(publicKey.Modulus),
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return EncodeBase64URL(digest[:]), nil
}
