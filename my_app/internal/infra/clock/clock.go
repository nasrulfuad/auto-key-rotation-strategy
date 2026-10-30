package clock

import (
	"time"

	"github.com/google/uuid"
)

type System struct{}

func New() System { return System{} }

func (System) Now() time.Time { return time.Now().UTC() }

type UUIDv4 struct{}

func NewUUIDGenerator() UUIDv4 { return UUIDv4{} }

func (UUIDv4) NewUUID() uuid.UUID { return uuid.New() }
