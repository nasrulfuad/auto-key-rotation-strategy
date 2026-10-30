package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type SigningKeyRepository struct {
	db               *sql.DB
	queries          store.Querier
	privateKeySealer port.PrivateKeySealer
}

var (
	_ port.SigningKeyReader = (*SigningKeyRepository)(nil)
	_ port.SigningKeyWriter = (*SigningKeyRepository)(nil)
)

func NewSigningKeyRepository(db *sql.DB, queries store.Querier, privateKeySealer port.PrivateKeySealer) *SigningKeyRepository {
	return &SigningKeyRepository{db: db, queries: queries, privateKeySealer: privateKeySealer}
}

func (r *SigningKeyRepository) Create(ctx context.Context, in port.SigningKeyCreate) (domain.SigningKey, error) {
	return r.create(ctx, r.queries, in)
}

func (r *SigningKeyRepository) Rotate(ctx context.Context, in port.SigningKeyRotate) (domain.SigningKey, error) {
	var nextKey domain.SigningKey
	err := withTransaction(ctx, r.db, func(queries *store.Queries) error {
		updated, err := queries.UpdateSigningKeyWindow(ctx, store.UpdateSigningKeyWindowParams{
			ID:          in.CurrentKeyID,
			ActiveUntil: in.CurrentKeyWindow.ActiveUntil,
			DeleteAfter: in.CurrentKeyWindow.DeleteAfter,
		})
		if err != nil {
			return domain.DB(err)
		}
		if updated == 0 {
			return domain.ErrSigningKeyNotFound
		}
		nextKey, err = r.create(ctx, queries, in.NextKey)
		return err
	})
	return nextKey, err
}

func (r *SigningKeyRepository) create(ctx context.Context, queries store.Querier, in port.SigningKeyCreate) (domain.SigningKey, error) {
	ciphertext, nonce, err := r.privateKeySealer.Seal(in.KeyPair.PrivateKey.Reveal())
	if err != nil {
		return domain.SigningKey{}, domain.Internal(err)
	}
	row, err := queries.InsertSigningKey(ctx, store.InsertSigningKeyParams{
		ID:                   in.ID,
		KeyID:                in.KeyPair.KeyID,
		Algorithm:            string(in.KeyPair.Algorithm),
		PublicKeyModulus:     in.KeyPair.PublicKey.Modulus,
		PublicKeyExponent:    in.KeyPair.PublicKey.Exponent,
		PrivateKeyCiphertext: ciphertext,
		PrivateKeyNonce:      nonce,
		IntroducedAt:         in.Window.IntroducedAt,
		ActiveFrom:           in.Window.ActiveFrom,
		ActiveUntil:          in.Window.ActiveUntil,
		DeleteAfter:          in.Window.DeleteAfter,
	})
	if err != nil {
		return domain.SigningKey{}, domain.DB(err)
	}
	return toSigningKey(store.FindActiveSigningKeyRow(row)), nil
}

func (r *SigningKeyRepository) FindLatest(ctx context.Context) (*domain.SigningKey, error) {
	row, err := r.queries.FindLatestSigningKey(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.DB(err)
	}
	key := toSigningKey(store.FindActiveSigningKeyRow(row))
	return &key, nil
}

func (r *SigningKeyRepository) FindActive(ctx context.Context, now time.Time) (domain.SigningKey, error) {
	row, err := r.queries.FindActiveSigningKey(ctx, now)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SigningKey{}, domain.ErrSigningKeyNotFound
	}
	if err != nil {
		return domain.SigningKey{}, domain.DB(err)
	}
	return toSigningKey(row), nil
}

func (r *SigningKeyRepository) FindPrivateKey(ctx context.Context, keyID string) (domain.Secret, error) {
	row, err := r.queries.FindSigningKeyPrivateKeyByKeyID(ctx, keyID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Secret{}, domain.ErrSigningKeyNotFound
	}
	if err != nil {
		return domain.Secret{}, domain.DB(err)
	}
	plaintext, err := r.privateKeySealer.Open(row.PrivateKeyCiphertext, row.PrivateKeyNonce)
	if err != nil {
		return domain.Secret{}, domain.Internal(err)
	}
	return domain.NewSecret(plaintext), nil
}

func (r *SigningKeyRepository) ListAllPublished(ctx context.Context, now time.Time) ([]domain.SigningKey, error) {
	rows, err := r.queries.ListAllPublishedSigningKeys(ctx, now)
	if err != nil {
		return nil, domain.DB(err)
	}
	keys := make([]domain.SigningKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, toSigningKey(store.FindActiveSigningKeyRow(row)))
	}
	return keys, nil
}

func (r *SigningKeyRepository) ListAll(ctx context.Context) ([]domain.SigningKey, error) {
	rows, err := r.queries.ListAllSigningKeys(ctx)
	if err != nil {
		return nil, domain.DB(err)
	}
	keys := make([]domain.SigningKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, toSigningKey(store.FindActiveSigningKeyRow(row)))
	}
	return keys, nil
}

func (r *SigningKeyRepository) DeleteExpired(ctx context.Context, now time.Time) ([]string, error) {
	keyIDs, err := r.queries.DeleteExpiredSigningKeys(ctx, now)
	if err != nil {
		return nil, domain.DB(err)
	}
	return keyIDs, nil
}

func toSigningKey(row store.FindActiveSigningKeyRow) domain.SigningKey {
	return domain.SigningKey{
		ID:        row.ID,
		KeyID:     row.KeyID,
		Algorithm: domain.SigningAlgorithm(row.Algorithm),
		PublicKey: domain.PublicKey{Modulus: row.PublicKeyModulus, Exponent: row.PublicKeyExponent},
		Window: domain.SigningKeyWindow{
			IntroducedAt: row.IntroducedAt.UTC(),
			ActiveFrom:   row.ActiveFrom.UTC(),
			ActiveUntil:  row.ActiveUntil.UTC(),
			DeleteAfter:  row.DeleteAfter.UTC(),
		},
		CreatedAt: row.CreatedAt.UTC(),
	}
}
