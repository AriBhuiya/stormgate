package stormgate

import (
	"github.com/aribhuiya/stormgate/internal/balancers"
	"github.com/aribhuiya/stormgate/internal/utils"
)

type Service struct {
	Config   utils.Service
	Balancer balancers.Balancer
}
