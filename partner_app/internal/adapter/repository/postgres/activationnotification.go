package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type ActivationNotificationRepository struct {
	queries store.Querier
}

var (
	_ port.ActivationNotificationReader = (*ActivationNotificationRepository)(nil)
	_ port.ActivationNotificationWriter = (*ActivationNotificationRepository)(nil)
)

func NewActivationNotificationRepository(queries store.Querier) *ActivationNotificationRepository {
	return &ActivationNotificationRepository{queries: queries}
}

func (r *ActivationNotificationRepository) ListAllDue(ctx context.Context, now time.Time, limit int) ([]domain.ActivationNotification, error) {
	rows, err := r.queries.ListAllDueActivationNotifications(ctx, store.ListAllDueActivationNotificationsParams{
		Now:      now,
		RowLimit: int32(limit),
	})
	if err != nil {
		return nil, domain.DB(err)
	}
	notifications := make([]domain.ActivationNotification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, domain.ActivationNotification{
			ID:            row.ID,
			Provider:      row.Provider,
			SubjectID:     row.SubjectID,
			EntitlementID: row.EntitlementID,
			Products:      row.Products,
			FreeTrial:     domain.FreeTrialStatus(row.FreeTrial),
			Event:         domain.ActivationNotificationEvent(row.Event),
			Status:        domain.ActivationNotificationStatus(row.Status),
			Attempts:      int(row.Attempts),
			NextAttemptAt: row.NextAttemptAt.UTC(),
			CreatedAt:     row.CreatedAt.UTC(),
		})
	}
	return notifications, nil
}

func (r *ActivationNotificationRepository) UpdateSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	if err := r.queries.UpdateActivationNotificationSent(ctx, store.UpdateActivationNotificationSentParams{ID: id, SentAt: sentAt}); err != nil {
		return domain.DB(err)
	}
	return nil
}

func (r *ActivationNotificationRepository) UpdateAttemptFailed(ctx context.Context, in port.ActivationNotificationAttempt) error {
	if err := r.queries.UpdateActivationNotificationAttemptFailed(ctx, store.UpdateActivationNotificationAttemptFailedParams{
		ID:            in.ID,
		Status:        string(in.Result.Status),
		Attempts:      int32(in.Result.Attempts),
		NextAttemptAt: in.Result.NextAttemptAt,
		LastError:     sql.NullString{String: in.LastError, Valid: in.LastError != ""},
	}); err != nil {
		return domain.DB(err)
	}
	return nil
}
