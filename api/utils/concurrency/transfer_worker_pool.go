package concurrency

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/extensions/attrs/maxworkersattr"
)

// RunInWorkerPool runs f(task) concurrently for each element in tasks,
// using up to maxWorkers workers. It waits for all tasks to complete
// and returns a joined error if any of them fail.
//
// If workers are set to 0 or 1, instead of spawning a traditional worker pool
// the tasks are run sequentially in order of task input. In this case, the
// returned error is the first error encountered, if any.
//
// For multiple workers, the worker pool does NOT cancel remaining tasks on first failure;
// all tasks run to completion. Context cancellation will stop idle
// workers but not interrupt running ones.
func RunInWorkerPool[T any](
	ctx context.Context,
	data datacontext.Context,
	tasks []T,
	f func(ctx context.Context, task T) error,
) error {
	if len(tasks) == 0 {
		return nil
	}
	maxWorkers, err := maxworkersattr.Get(data)
	if err != nil {
		return fmt.Errorf("failed to get max workers attribute: %w", err)
	}

	caller := getCaller()
	start := time.Now()
	logger := data.Logger().WithValues(
		"taksCount", len(tasks),
		"workers", maxWorkers,
		"caller", caller,
		"function", getFunctionName(f),
	)
	defer func() {
		logger.Debug("finished pool operation", "duration", time.Since(start))
	}()

	if maxWorkers == maxworkersattr.SingleWorker {
		logger.Debug("running tasks sequentially")
		for _, task := range tasks {
			if err := f(ctx, task); err != nil {
				return err
			}
		}
		return nil
	}

	if maxWorkers > uint(len(tasks)) {
		// this can happen if we run on a highly concurrent environment without a lot of tasks.
		// in this case we want to avoid spawning too many workers.
		maxWorkers = uint(len(tasks))
		logger.Debug("maxWorkers is greater than number of tasks, setting to number of tasks")
	}

	logger.Debug("starting worker pool")

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

// getCaller returns the caller of the function that calls it.
func getCaller() string {
	var pc [1]uintptr
	runtime.Callers(3, pc[:])
	f := runtime.FuncForPC(pc[0])
	if f == nil {
		return "caller unavailable"
	}
	return f.Name()
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
