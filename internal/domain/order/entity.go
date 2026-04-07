package order

import "time"

type Order struct {
	ID        int64
	VkOrderID string
	UserID    int64
	Product   string
	CreatedAt time.Time
}
