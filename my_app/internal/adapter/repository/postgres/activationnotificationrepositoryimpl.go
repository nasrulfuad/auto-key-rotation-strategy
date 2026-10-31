package postgres

import (
	"context"
	"database/sql"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/store"
)

type ActivationNotificationRepository struct {
	queries store.Querier
}

var (
	_ repository.ActivationNotificationReader = (*ActivationNotificationRepository)(nil)
	_ repository.ActivationNotificationWriter = (*ActivationNotificationRepository)(nil)
)

func NewActivationNotificationRepository(db *sql.DB, opts ...Option) *ActivationNotificationRepository {
	return &ActivationNotificationRepository{queries: newOptions(db, opts).querier}
}

func (r *ActivationNotificationRepository) Create(ctx context.Context, notification domain.ActivationNotification) (bool, error) {
	affected, err := r.queries.InsertActivationNotification(ctx, store.InsertActivationNotificationParams{
		ID:             notification.ID,
		NotificationID: notification.NotificationID,
		KeyID:          notification.KeyID,
		Issuer:         notification.Issuer,
		Event:          string(notification.Event),
		Provider:       notification.Provider,
		SubjectID:      notification.SubjectID,
		EntitlementID:  notification.EntitlementID,
		Products:       notification.Products,
		FreeTrial:      string(notification.FreeTrial),
		ReceivedAt:     notification.ReceivedAt,
	})
	if err != nil {
		return false, domain.DB(err)
	}
	return affected > 0, nil
}

func (r *ActivationNotificationRepository) ListAll(ctx context.Context, limit int) ([]domain.ActivationNotification, error) {
	rows, err := r.queries.ListAllActivationNotifications(ctx, int32(limit))
	if err != nil {
		return nil, domain.DB(err)
	}
	notifications := make([]domain.ActivationNotification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, domain.ActivationNotification{
			ID:             row.ID,
			NotificationID: row.NotificationID,
			KeyID:          row.KeyID,
			Issuer:         row.Issuer,
			Event:          domain.ActivationNotificationEvent(row.Event),
			Provider:       row.Provider,
			SubjectID:      row.SubjectID,
			EntitlementID:  row.EntitlementID,
			Products:       row.Products,
			FreeTrial:      domain.FreeTrialStatus(row.FreeTrial),
			ReceivedAt:     row.ReceivedAt.UTC(),
		})
	}
	return notifications, nil
}
