package errors

import (
	"errors"
	"syscall"
)

// IsRetryable checks whether a retry should be performed for a failed operation
func IsRetryable(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}
