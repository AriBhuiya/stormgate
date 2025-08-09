package balancers

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"math/rand"
	"net/http"
	"time"
)

type Random struct {
	service utils.Service
	seed    int32
}

func NewRandom(service *utils.Service, seed int32) (*Random, error) {
	if len(service.Backends) == 0 {
		err := errors.New("no available backends")
		return nil, err
	}
	return &Random{
		service: *service,
		seed:    seed,
	}, nil
}

func NewRandomAutoSeed(service *utils.Service) (*Random, error) {
	seed := time.Now().UnixNano()
	return NewRandom(service, int32(seed))
}

func (r *Random) PickBackend(request *http.Request) (string, error) {
	if len(r.service.Backends) == 0 {
		return "", errors.New("no healthy backends available")
	}
	idx := rand.Int() % len(r.service.Backends)
	return r.service.Backends[idx], nil
}

func (r *Random) SetHealthyBackends(healthyBackends []string) {
	if hasChanged := utils.HasBackendChanged(r.service.Backends, healthyBackends); !hasChanged {
		return
	}
	r.service.Backends = healthyBackends
}
