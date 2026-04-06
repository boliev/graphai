package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upsert(ctx context.Context, user *User) (*User, error) {
	if user == nil || user.PeerID == 0 || user.UserVKID == 0 {
		return nil, fmt.Errorf("invalid user")
	}

	return s.repo.Upsert(ctx, user)
}

func (s *Service) ReduceCredits(ctx context.Context, user *User) error {
	if user == nil || user.ID == 0 {
		return fmt.Errorf("invalid user")
	}

	return s.repo.ReduceCredits(ctx, user.ID)
}
