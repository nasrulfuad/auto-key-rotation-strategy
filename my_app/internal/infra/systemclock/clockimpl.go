package systemclock

import (
	"time"
)

type System struct{}

var _ Clock = System{}

func New() System { return System{} }

func (System) Now() time.Time { return time.Now().UTC() }
