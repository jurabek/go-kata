package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	logger := slog.Default()

	agg := NewAggregator(OrderSvc{}, ProfileSvc{}, WithTimeout(1*time.Second))
	res, err := agg.Aggregate(0)
	if err != nil {
		logger.Error("failed", slog.Any("err", err))
		return
	}

	logger.Info("result", slog.String("agg_res", res))
}

type Response map[string]any

func (r Response) String() string {
	var res string
	for key, val := range r {
		res = fmt.Sprintf("%s: %v", key, val)
		break
	}
	return res
}

type Fetcher interface {
	Fetch(ctx context.Context) (Response, error)
}

type OrderSvc struct{}

func (o OrderSvc) Fetch(ctx context.Context) (Response, error) {
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	// case <-time.After(2 * time.Second):
	// 	return map[string]any{"Orders": 5}, nil

	// second case
	case <-time.After(10 * time.Second):
		return map[string]any{"Orders": 5}, nil
	}
}

type ProfileSvc struct{}

func (p ProfileSvc) Fetch(ctx context.Context) (Response, error) {
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	// case <-time.After(2 * time.Second):
	// 	return map[string]any{"Name": "Alice"}, nil
	default:
		return map[string]any{"Name": "Alice"}, nil
	}
}

type Aggregator struct {
	orderFetcher   Fetcher
	profileFetcher Fetcher
	logger         *slog.Logger
	timeout        time.Duration
}

type AggregationOption func(*Aggregator)

func WithLogger(logger *slog.Logger) AggregationOption {
	return func(a *Aggregator) {
		a.logger = logger
	}
}

func WithTimeout(timeout time.Duration) AggregationOption {
	return func(a *Aggregator) {
		a.timeout = timeout
	}
}

func NewAggregator(orderFetcher, profileFetcher Fetcher, opts ...AggregationOption) *Aggregator {
	agg := &Aggregator{
		orderFetcher:   orderFetcher,
		profileFetcher: profileFetcher,
		logger:         slog.Default(),
		timeout:        5,
	}

	for _, opt := range opts {
		opt(agg)
	}

	return agg
}

func (a *Aggregator) Aggregate(id int) (string, error) {
	slog.Default().Info("timeout", slog.Any("timeout", a.timeout))

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	var userProfile Response
	errGroup.Go(func() error {
		profile, err := a.profileFetcher.Fetch(ctx)
		if err != nil {
			cancel()
			return err
		}
		userProfile = profile
		return nil
	})

	var order Response
	errGroup.Go(func() error {
		data, err := a.orderFetcher.Fetch(ctx)
		if err != nil {
			return err
		}
		order = data
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return "", err
	}

	return strings.Join([]string{userProfile.String(), order.String()}, "|"), nil
}
