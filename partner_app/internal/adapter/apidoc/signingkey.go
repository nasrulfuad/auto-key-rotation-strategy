package apidoc

import "time"

type JSONWebKey struct {
	KeyType   string `json:"kty" example:"RSA"`
	Exponent  string `json:"e" example:"AQAB"`
	Use       string `json:"use" example:"sig"`
	KeyID     string `json:"kid" example:"NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"`
	Algorithm string `json:"alg" example:"RS256" enums:"RS256,RS384,RS512"`
	Modulus   string `json:"n" example:"0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw"`
}

type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

type SigningKey struct {
	KeyID        string    `json:"key_id" example:"NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"`
	Algorithm    string    `json:"algorithm" example:"RS256"`
	Status       string    `json:"status" example:"active" enums:"pending,active,inactive,expired"`
	IsPublished  bool      `json:"is_published" example:"true"`
	IntroducedAt time.Time `json:"introduced_at" example:"2026-01-01T00:15:00Z"`
	ActiveFrom   time.Time `json:"active_from" example:"2026-01-01T00:20:00Z"`
	ActiveUntil  time.Time `json:"active_until" example:"2026-01-01T00:40:00Z"`
	DeleteAfter  time.Time `json:"delete_after" example:"2026-01-01T00:50:00Z"`
	CreatedAt    time.Time `json:"created_at" example:"2026-01-01T00:15:00Z"`
}

type SigningKeyList struct {
	Now  time.Time    `json:"now" example:"2026-01-01T00:25:00Z"`
	Data []SigningKey `json:"data"`
}

type PeerSigningKey struct {
	KeyID     string    `json:"key_id" example:"NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"`
	Algorithm string    `json:"algorithm" example:"RS256"`
	FetchedAt time.Time `json:"fetched_at" example:"2026-01-01T00:15:10Z"`
}

type PeerSigningKeyList struct {
	URL           string           `json:"url" example:"http://my_app:8081/.well-known/jwks.json"`
	LastFetchedAt *time.Time       `json:"last_fetched_at"`
	LastAttemptAt *time.Time       `json:"last_attempt_at"`
	LastError     *string          `json:"last_error"`
	Data          []PeerSigningKey `json:"data"`
}
