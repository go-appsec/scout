package scout

import (
	"errors"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	t.Parallel()

	errOne := errors.New("error one")
	errTwo := errors.New("error two")

	tests := []struct {
		name        string
		seq         iter.Seq2[string, error]
		wantResults []string
		wantErrs    []error
	}{
		{
			name: "empty_iterator",
			seq: func(yield func(string, error) bool) {
			},
			wantResults: nil,
			wantErrs:    nil,
		},
		{
			name: "all_successful",
			seq: func(yield func(string, error) bool) {
				yield("one", nil)
				yield("two", nil)
				yield("three", nil)
			},
			wantResults: []string{"one", "two", "three"},
			wantErrs:    nil,
		},
		{
			name: "all_errors",
			seq: func(yield func(string, error) bool) {
				yield("", errOne)
				yield("", errTwo)
			},
			wantResults: nil,
			wantErrs:    []error{errOne, errTwo},
		},
		{
			name: "mixed_results_and_errors",
			seq: func(yield func(string, error) bool) {
				yield("one", nil)
				yield("", errOne)
				yield("two", nil)
				yield("", errTwo)
			},
			wantResults: []string{"one", "two"},
			wantErrs:    []error{errOne, errTwo},
		},
		{
			name: "single_result",
			seq: func(yield func(string, error) bool) {
				yield("only", nil)
			},
			wantResults: []string{"only"},
			wantErrs:    nil,
		},
		{
			name: "single_error",
			seq: func(yield func(string, error) bool) {
				yield("", errOne)
			},
			wantResults: nil,
			wantErrs:    []error{errOne},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := Collect(tt.seq)

			assert.Equal(t, tt.wantResults, results)

			if tt.wantErrs == nil {
				assert.NoError(t, err)
			} else {
				for _, wantErr := range tt.wantErrs {
					assert.ErrorIs(t, err, wantErr)
				}
			}
		})
	}
}
