package systemclock

import "github.com/google/uuid"

type UUIDv4 struct{}

var _ UUIDGenerator = UUIDv4{}

func NewUUIDGenerator() UUIDv4 { return UUIDv4{} }

func (UUIDv4) NewUUID() uuid.UUID { return uuid.New() }
