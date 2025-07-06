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
	index := (r.counter.Add(1) - 1) % r.n
	return r.service.Backends[index], nil
}
