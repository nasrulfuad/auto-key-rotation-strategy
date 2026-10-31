package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/gateway"
)

const maximumResponseBytes = 4096

type ActivationNotificationClient struct {
	url    string
	client *http.Client
}

var _ gateway.ActivationNotificationSender = (*ActivationNotificationClient)(nil)

func NewActivationNotificationClient(url string, opts ...Option) *ActivationNotificationClient {
	client := newOptions(opts).httpClient
	return &ActivationNotificationClient{url: url, client: client}
}

type activationNotificationRequest struct {
	NotificationID string `json:"notification_id"`
	Event          string `json:"event"`
	Provider       string `json:"provider"`
	SubjectID      string `json:"subject_id"`
	EntitlementID  string `json:"entitlement_id"`
	Products       string `json:"products"`
	FreeTrial      string `json:"free_trial"`
}

func (c *ActivationNotificationClient) Send(ctx context.Context, in gateway.ActivationNotificationDelivery) error {
	body, err := json.Marshal(activationNotificationRequest{
		NotificationID: in.Notification.ID.String(),
		Event:          string(in.Notification.Event),
		Provider:       in.Notification.Provider,
		SubjectID:      in.Notification.SubjectID,
		EntitlementID:  in.Notification.EntitlementID,
		Products:       in.Notification.Products,
		FreeTrial:      string(in.Notification.FreeTrial),
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+in.Token)
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, maximumResponseBytes))
	return fmt.Errorf("activation notification endpoint responded with status %d: %s", response.StatusCode, bytes.TrimSpace(responseBody))
}
