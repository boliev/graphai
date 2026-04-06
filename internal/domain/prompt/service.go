package prompt

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, prompt *Prompt) error {
	return s.repository.Create(ctx, prompt)
}
