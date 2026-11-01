package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]Result[R], error) {
	// edge cases
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	inLen := len(in)
	results := make([]Result[R], inLen)
	if inLen == 0 {
		return results, nil
	}

	// worker pool
	if workers > inLen {
		workers = inLen
	}

	var wg sync.WaitGroup
	var idx atomic.Int64
	for range workers {
		wg.Go(func() {
			for {
				if err := ctx.Err(); err != nil {
					return
				}

				i := idx.Add(1) - 1
				if i >= int64(inLen) {
					return
				}

				res, err := fn(ctx, in[i])
				if err != nil {
					results[i] = Result[R]{Err: err}
				} else {
					results[i] = Result[R]{Value: res}
				}
			}
		})
	}

	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
