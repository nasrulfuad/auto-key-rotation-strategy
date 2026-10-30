package systemclock

import "time"

type Clock interface {
	Now() time.Time
}
