package requests

import (
	"context"

	"github.com/superkruger/nostr_app_data/app/domain"
)

type Service interface {
	Add(ctx context.Context, request domain.Request) error
	Remove(ctx context.Context, id string) error
	Find(ctx context.Context, event domain.Event) ([]domain.Request, error)
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

func (s *service) Add(ctx context.Context, request domain.Request) error {
	return s.repo.add(ctx, request)
}

func (s *service) Remove(ctx context.Context, id string) error {
	return s.repo.remove(ctx, id)
}

func (s *service) Find(ctx context.Context, event domain.Event) ([]domain.Request, error) {
	return s.repo.findForEvent(ctx, event)
}
