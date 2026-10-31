package systemclock

import "time"

// Clock tells the current time. A use case calls Now once per execution and passes the value down.
type Clock interface {
	// Now returns the current time in UTC.
	Now() time.Time
}
