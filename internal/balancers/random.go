package balancers

import "github.com/aribhuiya/stormgate/internal/utils"

type Random struct {
	service utils.Service
	seed    int32
}

func NewRandom(service utils.Service, seed int32) *Random {
	return &Random{
		service: service,
		seed:    seed,
	}
}
