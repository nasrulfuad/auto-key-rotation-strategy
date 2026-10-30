-- +goose Up
CREATE TABLE signing_key (
    id UUID NOT NULL,
    key_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    public_key_modulus BYTEA NOT NULL,
    public_key_exponent BYTEA NOT NULL,
    private_key_ciphertext BYTEA NOT NULL,
    private_key_nonce BYTEA NOT NULL,
    introduced_at TIMESTAMPTZ NOT NULL,
    active_from TIMESTAMPTZ NOT NULL,
    active_until TIMESTAMPTZ NOT NULL,
    delete_after TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT signing_key_pkey PRIMARY KEY (id),
    CONSTRAINT signing_key_key_id_key UNIQUE (key_id),
    CONSTRAINT signing_key_active_from_key UNIQUE (active_from),
    CONSTRAINT signing_key_algorithm_chk CHECK (algorithm IN ('RS256', 'RS384', 'RS512')),
    CONSTRAINT signing_key_window_chk CHECK (introduced_at <= active_from AND active_from < active_until AND active_until <= delete_after)
);
CREATE INDEX signing_key_delete_after_idx ON signing_key (delete_after);

CREATE TABLE activation_notification (
    id UUID NOT NULL,
    notification_id UUID NOT NULL,
    key_id TEXT NOT NULL,
    issuer TEXT NOT NULL,
    event TEXT NOT NULL,
    provider TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    entitlement_id TEXT NOT NULL,
    products TEXT NOT NULL,
    free_trial TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT activation_notification_pkey PRIMARY KEY (id),
    CONSTRAINT activation_notification_notification_id_key UNIQUE (notification_id)
);
CREATE INDEX activation_notification_received_at_idx ON activation_notification (received_at);
