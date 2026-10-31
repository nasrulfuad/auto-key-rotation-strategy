package systemclock

import "github.com/google/uuid"

// UUIDGenerator generates row identifiers and token IDs.
type UUIDGenerator interface {
	// NewUUID returns a new random version 4 UUID.
	NewUUID() uuid.UUID
}
