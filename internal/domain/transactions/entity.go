package transactions

import "time"

const (
	OPERATION_TYPE_DEBIT      = "debit"
	OPERATION_TYPE_CREDIT     = "credit"
	OPERATION_TYPE_FREE_USAGE = "free_usage"
)

type Transaction struct {
	ID            int64
	UserID        int64
	Prompt        string
	OperationType string
	Amount        int64
	CreatedAt     time.Time
}
