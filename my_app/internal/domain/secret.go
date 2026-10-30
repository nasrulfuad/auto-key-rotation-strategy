package domain

import (
	"errors"
	"log/slog"
)

const redacted = "[REDACTED]"

var errSecretNotSerializable = errors.New("secret values cannot be serialized")

type Secret struct {
	value []byte
}

func NewSecret(value []byte) Secret { return Secret{value: value} }

func (s Secret) Reveal() []byte { return s.value }

func (s Secret) IsEmpty() bool { return len(s.value) == 0 }

func (Secret) String() string { return redacted }

func (Secret) GoString() string { return redacted }

func (Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

func (Secret) MarshalJSON() ([]byte, error) { return nil, errSecretNotSerializable }

func (Secret) MarshalText() ([]byte, error) { return nil, errSecretNotSerializable }
