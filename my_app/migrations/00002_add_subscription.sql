-- +goose Up
CREATE TABLE subscription (
    id UUID NOT NULL,
    subject_id TEXT NOT NULL,
    entitlement_id TEXT NOT NULL,
    products TEXT NOT NULL,
    free_trial TEXT NOT NULL,
    status TEXT NOT NULL,
    signed_with_key_id TEXT NOT NULL,
    activated_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT subscription_pkey PRIMARY KEY (id),
    CONSTRAINT subscription_subject_id_entitlement_id_key UNIQUE (subject_id, entitlement_id),
    CONSTRAINT subscription_status_chk CHECK (status IN ('ACTIVE', 'TERMINATED')),
    CONSTRAINT subscription_terminated_at_chk CHECK ((status = 'TERMINATED') = (terminated_at IS NOT NULL))
);
CREATE INDEX subscription_updated_at_idx ON subscription (updated_at);
