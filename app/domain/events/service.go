package events

import (
	"context"
	"fmt"
	"time"

	"github.com/superkruger/nostr_app_data/app/domain"
)

const (
	ttlSecondsRegular     = 3600 * 24 * 7 * time.Second
	ttlSecondsReplaceable = 3600 * 24 * 30 * time.Second
	ttlSecondsAddressable = 3600 * 24 * 30 * time.Second
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
	now := time.Now()
	spam, err := s.repo.isSpam(ctx, dbEvent, now)
	if err != nil {
		return err
	}
	if spam {
		return fmt.Errorf("event is spam")
	}
	if dbEvent.IsRegular() {
		dbEvent.ExpireAt = now.Add(ttlSecondsRegular)
		return s.repo.add(ctx, dbEvent)
	}
	if dbEvent.IsReplaceable() {
		dbEvent.ExpireAt = now.Add(ttlSecondsReplaceable)
		return s.repo.replace(ctx, dbEvent)
	}
	if dbEvent.IsAddressable() {
		dbEvent.ExpireAt = now.Add(ttlSecondsAddressable)
		return s.repo.replaceAddressable(ctx, dbEvent)
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
