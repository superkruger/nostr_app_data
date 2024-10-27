package events

import (
	"context"
	"fmt"
)

type Service interface {
	Add(ctx context.Context, event Event) error
	Remove(ctx context.Context, id string) error
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

func (s *service) Add(ctx context.Context, event Event) error {
	valid, err := event.CheckSignature()
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid signature")
	}
	return s.repo.add(ctx, event.toDB())
}

func (s *service) Remove(ctx context.Context, id string) error {
	return s.repo.remove(ctx, id)
}
