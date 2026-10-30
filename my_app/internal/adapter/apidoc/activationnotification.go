package apidoc

import "time"

type ActivationNotification struct {
	NotificationID string    `json:"notification_id" example:"6a202257-04fd-4b17-93ce-370b9b99ef12"`
	KeyID          string    `json:"key_id" example:"zonIzOJt60M94v1hq7rH2ddIYuqjTigBsXwROMiPV7M"`
	Issuer         string    `json:"issuer" example:"qa.partner-app.local"`
	Event          string    `json:"event" example:"entitlement_upserted"`
	Provider       string    `json:"provider" example:"my_app"`
	SubjectID      string    `json:"subject_id" example:"SUB001"`
	EntitlementID  string    `json:"entitlement_id" example:"ENT001"`
	Products       string    `json:"products" example:"com.partner.product1"`
	FreeTrial      string    `json:"free_trial" example:"NONE"`
	ReceivedAt     time.Time `json:"received_at" example:"2026-01-01T00:25:05Z"`
}

type ActivationNotificationList struct {
	Data []ActivationNotification `json:"data"`
}

type ReceiveActivationNotificationRequest struct {
	NotificationID string `json:"notification_id" example:"6a202257-04fd-4b17-93ce-370b9b99ef12"`
	Event          string `json:"event" example:"entitlement_upserted" enums:"entitlement_upserted,entitlement_terminated"`
	Provider       string `json:"provider" example:"my_app"`
	SubjectID      string `json:"subject_id" example:"SUB001"`
	EntitlementID  string `json:"entitlement_id" example:"ENT001"`
	Products       string `json:"products" example:"com.partner.product1"`
	FreeTrial      string `json:"free_trial" example:"NONE"`
}
