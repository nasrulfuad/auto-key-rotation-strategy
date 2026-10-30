-- +goose Up
ALTER TABLE entitlement ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE entitlement ADD COLUMN terminated_at TIMESTAMPTZ;
ALTER TABLE entitlement ADD CONSTRAINT entitlement_status_chk CHECK (status IN ('ACTIVE', 'TERMINATED'));
ALTER TABLE entitlement ADD CONSTRAINT entitlement_terminated_at_chk CHECK ((status = 'TERMINATED') = (terminated_at IS NOT NULL));
ALTER TABLE activation_notification DROP CONSTRAINT activation_notification_event_chk;
ALTER TABLE activation_notification ADD CONSTRAINT activation_notification_event_chk CHECK (event IN ('entitlement_upserted', 'entitlement_terminated'));
