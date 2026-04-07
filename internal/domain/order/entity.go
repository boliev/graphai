package order

import "time"

type Order struct {
	ID        int64
	VkOrderID int64
	UserID    int64
	Product   string
	CreatedAt time.Time
}
