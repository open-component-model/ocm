package transfer

import (
	"context"
	"errors"
	"sync"
)

// runWorkerPool runs f(task) concurrently for each element in tasks,
// using up to maxWorkers workers. It waits for all tasks to complete
// and returns a joined error if any of them fail.
//
// The worker pool does NOT cancel remaining tasks on first failure;
// all tasks run to completion. Context cancellation will stop idle
// workers but not interrupt running ones.
func runWorkerPool[T any](
	ctx context.Context,
	tasks []T,
	maxWorkers uint,
	f func(ctx context.Context, task T) error,
) error {
	if len(tasks) == 0 {
		return nil
	}
	if maxWorkers == 0 {
		maxWorkers = 1
	}

	taskCh := make(chan T, len(tasks))
	errCh := make(chan error, len(tasks))

	var wg sync.WaitGroup

	// start worker pool
	for i := uint(0); i < maxWorkers; i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskCh:
					if !ok {
						return
					}
					if err := f(ctx, task); err != nil {
						errCh <- err
					}
				}
			}
		})
	}

	// enqueue all tasks
	for _, t := range tasks {
		taskCh <- t
	}

	close(taskCh)
	wg.Wait()
	close(errCh)

	var combined error
	for err := range errCh {
		combined = errors.Join(combined, err)
	}
	return combined
}
