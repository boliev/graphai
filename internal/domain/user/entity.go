package user

import "time"

type User struct {
	ID         int64
	UserVKID   int64
	PeerID     int64
	Credits    int64
	LastAction time.Time
	LastNotify time.Time
	CreatedAt  time.Time
}

func (u *User) HasBalance() bool {
	return u.Credits > 0
}
