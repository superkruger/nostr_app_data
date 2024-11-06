package events

import (
	"context"
	"fmt"

	"github.com/superkruger/nostr_app_data/app/domain"
)

type Service interface {
	Add(ctx context.Context, event domain.Event) error
	Remove(ctx context.Context, id string) error
	FindForRequest(ctx context.Context, req domain.Request) ([]domain.Event, error)
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

func (s *service) Add(ctx context.Context, event domain.Event) error {
	valid, err := event.CheckSignature()
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid signature")
	}
	dbEvent := event.ToDB()
	if event.IsRegular() {
		return s.repo.add(ctx, dbEvent)
	}
	if event.IsReplaceable() {
		if err := s.repo.removeReplaceable(ctx, event.PubKey, event.Kind); err != nil {
			return err
		}
		return s.repo.add(ctx, dbEvent)
	}
	if event.IsAddressable() {
		if err := s.repo.removeAddressable(ctx, event.PubKey, event.Kind, dbEvent.Tags["d"]); err != nil {
			return err
		}
		return s.repo.add(ctx, dbEvent)
	}
	return nil
}

func (s *service) Remove(ctx context.Context, id string) error {
	return s.repo.remove(ctx, id)
}

func (s *service) FindForRequest(ctx context.Context, req domain.Request) ([]domain.Event, error) {
	dbEvents, err := s.repo.findForRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Event, 0, len(dbEvents))
	for _, dbEvent := range dbEvents {
		res = append(res, dbEvent.ToJson())
	}
	return res, nil
}
