package routing_strategy

import (
	"fmt"
	"github.com/aribhuiya/stormgate/internal/utils"
)

type RoutingStrategy interface {
	Route(prefixPath *string) (*RouteEntry, error)
}

type RouteEntry struct {
	Path    string
	Service *utils.Services
}

func CreateRoutingStrategy(name string, services *[]utils.Services) RoutingStrategy {
	switch name {
	case "hybrid":
		return NewHttpHybridRouting(services)
	case "simple", "":
		return NewSimpleRouting(services)
	default:
		panic(fmt.Sprintf("unsupported routing strategy- Use hybrid or simple or blank: %s", name))
	}
}
