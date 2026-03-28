package transactions

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, tx *Transaction) error {
	return s.repository.Create(ctx, tx)
}
