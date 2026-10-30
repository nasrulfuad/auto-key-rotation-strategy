package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type EntitlementRepository struct {
	db      *sql.DB
	queries store.Querier
}

var (
	_ port.EntitlementReader = (*EntitlementRepository)(nil)
	_ port.EntitlementWriter = (*EntitlementRepository)(nil)
)

func NewEntitlementRepository(db *sql.DB, queries store.Querier) *EntitlementRepository {
	return &EntitlementRepository{db: db, queries: queries}
}

func (r *EntitlementRepository) Upsert(ctx context.Context, in port.EntitlementUpsert) (domain.Entitlement, error) {
	var entitlement domain.Entitlement
	err := withTransaction(ctx, r.db, func(queries *store.Queries) error {
		row, err := queries.UpsertEntitlement(ctx, store.UpsertEntitlementParams{
			ID:            in.ID,
			Provider:      in.Change.Provider,
			SubjectID:     in.Change.SubjectID,
			EntitlementID: in.Change.EntitlementID,
			Products:      in.Change.Products,
			FreeTrial:     string(in.Change.FreeTrial),
			LastUpdatedAt: in.Change.LastUpdatedAt,
			Now:           in.Now,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrEntitlementStaleUpdate
		}
		if err != nil {
			return domain.DB(err)
		}
		if err := queries.InsertActivationNotification(ctx, store.InsertActivationNotificationParams{
			ID:               in.ActivationNotificationID,
			EntitlementRowID: row.ID,
			Event:            string(domain.ActivationNotificationEventEntitlementUpserted),
			Now:              in.Now,
		}); err != nil {
			return domain.DB(err)
		}
		entitlement = toEntitlement(store.FindEntitlementBySubjectAndEntitlementRow(row))
		return nil
	})
	return entitlement, err
}

func (r *EntitlementRepository) Terminate(ctx context.Context, in port.EntitlementTerminate) (domain.Entitlement, error) {
	var entitlement domain.Entitlement
	err := withTransaction(ctx, r.db, func(queries *store.Queries) error {
		row, err := queries.UpdateEntitlementTermination(ctx, store.UpdateEntitlementTerminationParams{
			Now:           in.Now,
			LastUpdatedAt: in.Termination.LastUpdatedAt,
			Provider:      in.Termination.Provider,
			SubjectID:     in.Termination.SubjectID,
			EntitlementID: in.Termination.EntitlementID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return r.toTerminationMiss(ctx, queries, in.Termination)
		}
		if err != nil {
			return domain.DB(err)
		}
		if err := queries.InsertActivationNotification(ctx, store.InsertActivationNotificationParams{
			ID:               in.ActivationNotificationID,
			EntitlementRowID: row.ID,
			Event:            string(domain.ActivationNotificationEventEntitlementTerminated),
			Now:              in.Now,
		}); err != nil {
			return domain.DB(err)
		}
		entitlement = toEntitlement(store.FindEntitlementBySubjectAndEntitlementRow(row))
		return nil
	})
	return entitlement, err
}

func (r *EntitlementRepository) toTerminationMiss(ctx context.Context, queries *store.Queries, termination domain.EntitlementTermination) error {
	found, err := queries.EntitlementExists(ctx, store.EntitlementExistsParams{
		Provider:      termination.Provider,
		SubjectID:     termination.SubjectID,
		EntitlementID: termination.EntitlementID,
	})
	if err != nil {
		return domain.DB(err)
	}
	if !found {
		return domain.ErrEntitlementNotFound
	}
	return domain.ErrEntitlementStaleUpdate
}

func (r *EntitlementRepository) FindBySubjectAndEntitlement(ctx context.Context, provider, subjectID, entitlementID string) (domain.Entitlement, error) {
	row, err := r.queries.FindEntitlementBySubjectAndEntitlement(ctx, store.FindEntitlementBySubjectAndEntitlementParams{
		Provider:      provider,
		SubjectID:     subjectID,
		EntitlementID: entitlementID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Entitlement{}, domain.ErrEntitlementNotFound
	}
	if err != nil {
		return domain.Entitlement{}, domain.DB(err)
	}
	return toEntitlement(row), nil
}

func (r *EntitlementRepository) ListAllBySubject(ctx context.Context, provider, subjectID string) ([]domain.Entitlement, error) {
	rows, err := r.queries.ListAllEntitlementsBySubject(ctx, store.ListAllEntitlementsBySubjectParams{Provider: provider, SubjectID: subjectID})
	if err != nil {
		return nil, domain.DB(err)
	}
	entitlements := make([]domain.Entitlement, 0, len(rows))
	for _, row := range rows {
		entitlements = append(entitlements, toEntitlement(store.FindEntitlementBySubjectAndEntitlementRow(row)))
	}
	return entitlements, nil
}

func toEntitlement(row store.FindEntitlementBySubjectAndEntitlementRow) domain.Entitlement {
	return domain.Entitlement{
		ID:            row.ID,
		Provider:      row.Provider,
		SubjectID:     row.SubjectID,
		EntitlementID: row.EntitlementID,
		Products:      row.Products,
		FreeTrial:     domain.FreeTrialStatus(row.FreeTrial),
		Status:        domain.EntitlementStatus(row.Status),
		TerminatedAt:  toUTC(row.TerminatedAt),
		LastUpdatedAt: row.LastUpdatedAt.UTC(),
		CreatedAt:     row.CreatedAt.UTC(),
		UpdatedAt:     row.UpdatedAt.UTC(),
	}
}

func toUTC(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
