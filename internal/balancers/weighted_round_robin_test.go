package balancers

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"testing"
)

func TestWeightedRoundRobin_PickBackend(t *testing.T) {
	tests := []struct {
		name        string
		backends    []string
		weights     []int32
		totalWeight uint64
		expectedSeq []string // Expected output for successive calls
		expectErr   bool
	}{
		{
			name:        "single backend",
			backends:    []string{"A"},
			weights:     []int32{1},
			totalWeight: 1,
			expectedSeq: []string{"A", "A", "A"},
		},
		{
			name:        "two backends equal weight",
			backends:    []string{"A", "B"},
			weights:     []int32{1, 1},
			totalWeight: 2,
			expectedSeq: []string{"A", "B", "A", "B"},
		},
		{
			name:        "three backends with varying weights",
			backends:    []string{"A", "B", "C"},
			weights:     []int32{3, 1, 2},
			totalWeight: 6,
			expectedSeq: []string{"A", "A", "A", "B", "C", "C", "A"}, // wrapping around
		},
		{
			name:        "no backends",
			backends:    []string{},
			weights:     []int32{},
			totalWeight: 0,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightedRoundRobin{
				service:     &utils.Service{Backends: tt.backends},
				weights:     tt.weights,
				totalWeight: tt.totalWeight,
			}
			for i, want := range tt.expectedSeq {
				got, err := w.PickBackend(nil)
				if (err != nil) != tt.expectErr {
					t.Fatalf("PickBackend() error = %v, wantErr %v", err, tt.expectErr)
				}
				if err == nil && got != want {
					t.Errorf("PickBackend() call %d = %v, want %v", i+1, got, want)
				}
			}
		})
	}
}
