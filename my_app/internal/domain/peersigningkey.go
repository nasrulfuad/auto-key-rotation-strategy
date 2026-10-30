package domain

import "time"

type PeerSigningKey struct {
	KeyID     string
	Algorithm SigningAlgorithm
	PublicKey PublicKey
	FetchedAt time.Time
}

type PeerKeySetStatus struct {
	URL           string
	LastFetchedAt time.Time
	LastAttemptAt time.Time
	LastError     string
}
