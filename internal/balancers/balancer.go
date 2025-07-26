package balancers

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/balancers/consistent_hash"
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
)

type Balancer interface {
	// PickBackend returning a string as copied should be fine as Go creates a pointer to the raw data under the hood
	PickBackend(request *http.Request) (string, error)
}

func Create(name string, service *utils.Service) (Balancer, error) {
	switch name {
	case "round_robin":
		return NewRoundRobin(service)
	case "random":
		return NewRandomAutoSeed(service)
	case "weighted_round_robin":
		return NewWeightedRoundRobin(service)
	case "consistent_hash":
		return consistent_hash.NewHashModulo(service)

	}

	return nil, errors.New("unknown balancer type " + name)
}
