package systemclock

import "github.com/google/uuid"

type UUIDGenerator interface {
	NewUUID() uuid.UUID
}
