package user

import "time"

type User struct {
	ID         int64
	ChatID     int64
	Username   string
	FirstName  string
	LastName   string
	Balance    int64
	LastAction time.Time
	LastNotify time.Time
	CreatedAt  time.Time
}
