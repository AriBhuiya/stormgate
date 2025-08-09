package balancers

import (
	"errors"
	"fmt"
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"sync/atomic"
)

type WeightedRoundRobin struct {
	counter     atomic.Uint64 // for counting the total requests. thread-safe at CPU level
	service     *utils.Service
	n           uint64
	weights     []int32
	totalWeight uint64
	allBackends []string
	allWeights  []int32
}

func validateWeights(backends []string, weights []int32) error {
	if len(backends) != len(weights) {
		return fmt.Errorf("number of weights (%d) does not match number of backends (%d)", len(weights), len(backends))
	}

	for i, w := range weights {
		if w == 0 {
			return fmt.Errorf("weight at index %d is zero — must be a positive integer", i)
		}
	}
	return nil
}

func (w *WeightedRoundRobin) PickBackend(*http.Request) (string, error) {
	if len(w.service.Backends) == 0 || len(w.weights) != len(w.service.Backends) {
		return "", errors.New("no healthy backends")
	}

	index := (w.counter.Add(1) - 1) % w.totalWeight

	accum := uint64(0)
	for i, weight := range w.weights {
		accum += uint64(weight)
		if index < accum {
			return w.service.Backends[i], nil
		}
	}

	return "", errors.New("unreachable state")
}

func normalizeWeights(weights []int32) []int32 {
	gcd := func(a, b int32) int32 {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	g := weights[0]
	for _, w := range weights[1:] {
		g = gcd(g, w)
	}

	if g == 1 {
		return weights
	}

	norm := make([]int32, len(weights))
	for i, w := range weights {
		norm[i] = w / g
	}
	return norm
}

func NewWeightedRoundRobin(service *utils.Service) (*WeightedRoundRobin, error) {
	if len(service.Backends) == 0 {
		err := errors.New("no available backends")
		return nil, err
	}

	rawWeights, ok := service.StrategyConfig["weights"].([]interface{})
	if !ok {
		return nil, errors.New("weights missing or malformed")
	}
	weights := make([]int32, len(rawWeights))
	for i, w := range rawWeights {
		intVal, ok := w.(int)
		if !ok {
			return nil, fmt.Errorf("weight at index %d is not an int", i)
		}
		weights[i] = int32(intVal)
	}

	err := validateWeights(service.Backends, weights)
	if err != nil {
		return nil, err
	}

	normalized := normalizeWeights(weights)
	totalWeight := uint64(0)
	for _, weight := range normalized {
		totalWeight += uint64(weight)
	}

	return &WeightedRoundRobin{
		counter:     atomic.Uint64{},
		service:     service,
		n:           uint64(len(service.Backends)),
		weights:     normalized,
		totalWeight: totalWeight,
		allBackends: append([]string(nil), service.Backends...),
		allWeights:  append([]int32(nil), normalized...),
	}, nil
}

func (w *WeightedRoundRobin) SetHealthyBackends(healthyBackends []string) {
	if hasChanged := utils.HasBackendChanged(w.service.Backends, healthyBackends); !hasChanged {
		return
	}
	if len(healthyBackends) == 0 {
		w.service.Backends = nil
		w.weights = nil
		w.totalWeight = 0
		return
	}

	healthySet := make(map[string]bool)
	for _, h := range healthyBackends {
		healthySet[h] = true
	}

	newBackends := make([]string, 0, len(healthyBackends))
	newWeights := make([]int32, 0, len(healthyBackends))

	for i, backend := range w.allBackends {
		if healthySet[backend] {
			newBackends = append(newBackends, backend)
			newWeights = append(newWeights, w.allWeights[i])
		}
	}

	norm := normalizeWeights(newWeights)
	total := uint64(0)
	for _, w := range norm {
		total += uint64(w)
	}

	w.service.Backends = newBackends
	w.weights = norm
	w.totalWeight = total
}
