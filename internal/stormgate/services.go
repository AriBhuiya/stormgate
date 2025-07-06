package stormgate

import (
	"github.com/aribhuiya/stormgate/internal/balancers"
	"github.com/aribhuiya/stormgate/internal/utils"
)

type service struct {
	config   utils.Service
	balancer balancers.Balancer
}
