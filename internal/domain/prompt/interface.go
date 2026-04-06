package prompt

import "context"

type Repository interface {
	Create(ctx context.Context, ptompt *Prompt) error
}
