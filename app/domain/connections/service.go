package connections

import (
	"context"
	"time"
)

type Service interface {
	Add(ctx context.Context, id string, at time.Time) error
	Remove(ctx context.Context, id string) error
	All(ctx context.Context) ([]connection, error)
}

type service struct {
	repo Repository
}

func NewService(opts ...func(svc *service)) Service {
	svc := &service{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func WithRepo(repo Repository) func(svc *service) {
	return func(svc *service) {
		svc.repo = repo
	}
}

func (s *service) Add(ctx context.Context, id string, at time.Time) error {
	return s.repo.add(ctx, connection{
		ID:        id,
		CreatedAt: at,
	})
}

func (s *service) Remove(ctx context.Context, id string) error {
	return s.repo.remove(ctx, id)
}

func (s *service) All(ctx context.Context) ([]connection, error) {
	return s.repo.all(ctx)
}
