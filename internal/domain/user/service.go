package user

import (
	"context"
	"errors"
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
	if user == nil || user.UserVKID == 0 {
		return nil, errors.New("invalid user")
	}

	return s.repo.Upsert(ctx, user)
}

func (s *Service) ReduceCredits(ctx context.Context, user *User) error {
	if user == nil || user.ID == 0 {
		return errors.New("invalid user")
	}

	return s.repo.ReduceCredits(ctx, user.ID)
}

func (s *Service) IncreaseCreditsTx(ctx context.Context, tx pgx.Tx, userID, credits int64) error {
	return s.repo.IncreaseCreditsTx(ctx, tx, userID, credits)
}

func (s *Service) FindByVKID(ctx context.Context, id int64) (*User, error) {
	return s.repo.FindByVKID(ctx, id)
}

func (s *Service) FindByVkIDOrUpsert(ctx context.Context, vkUserID int64) (*User, error) {
	usr, err := s.FindByVKID(ctx, vkUserID)
	if err != nil {
		return nil, fmt.Errorf("cannot get user: %w", err)
	}

	if usr == nil {
		usr, err = s.Upsert(ctx, &User{
			UserVKID: vkUserID,
			PeerID:   0,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot create user: %w", err)
		}
	}

	return usr, nil
}
