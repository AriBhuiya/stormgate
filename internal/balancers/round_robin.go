package balancers

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"sync/atomic"
)

type RoundRobin struct {
	counter atomic.Uint64 // for counting the total requests. thread-safe at CPU level
	service *utils.Service
	n       uint64
}

func NewRoundRobin(service *utils.Service) (*RoundRobin, error) {
	if len(service.Backends) == 0 {
		err := errors.New("no available backends")
		return nil, err
	}
	return &RoundRobin{
		counter: atomic.Uint64{},
		service: service,
		n:       uint64(len(service.Backends)),
	}, nil
}

func (r *RoundRobin) PickBackend(*http.Request) (string, error) {
	if r.n == 0 {
		return "", errors.New("no healthy backends available")
	}
	index := (r.counter.Add(1) - 1) % r.n
	return r.service.Backends[index], nil
}

func (r *RoundRobin) SetHealthyBackends(healthyBackends []string) {
	if hasChanged := utils.HasBackendChanged(r.service.Backends, healthyBackends); !hasChanged {
		return
	}
	r.service.Backends = healthyBackends
	r.n = uint64(len(healthyBackends))
}
