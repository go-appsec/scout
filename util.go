package scout

import (
	"errors"
	"iter"
)

// Collect iterates over all results, collecting results without error.
// Returned are the collected non-error results, if any errors occurred they will be joined into a non-nil error result.
// Partial results are returned alongside errors.
func Collect[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var results []T
	var errs []error
	for v, err := range seq {
		if err != nil {
			errs = append(errs, err)
		} else {
			results = append(results, v)
		}
	}
	return results, errors.Join(errs...)
}
