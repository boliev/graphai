package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (s *Service) IncreaseCreditsTx(ctx context.Context, tx pgx.Tx, userID int64, credits int64) error {
	return s.repo.IncreaseCreditsTx(ctx, tx, userID, credits)
}

func (s *Service) FindByVKID(ctx context.Context, id int64) (*User, error) {
	return s.repo.FindByVKID(ctx, id)
}
