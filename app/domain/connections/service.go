package connections

import (
	"context"
	"time"
)

type Service interface {
	AddConnection(ctx context.Context, id string, at time.Time) error
	RemoveConnection(ctx context.Context, id string) error
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

func (s *service) AddConnection(ctx context.Context, id string, at time.Time) error {
	return s.repo.add(ctx, connection{
		ID:        id,
		CreatedAt: at,
	})
}

func (s *service) RemoveConnection(ctx context.Context, id string) error {
	return s.repo.remove(ctx, id)
}
