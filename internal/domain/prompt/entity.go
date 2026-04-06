package prompt

import "time"

type Prompt struct {
	ID        int64
	UserID    int64
	Prompt    string
	CreatedAt time.Time
}
