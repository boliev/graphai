package user

import "time"

type User struct {
	ID         int64
	UserVKID   int64
	PeerID     int64
	Balance    int64
	FreeUsages int64
	LastAction time.Time
	LastNotify time.Time
	CreatedAt  time.Time
}
