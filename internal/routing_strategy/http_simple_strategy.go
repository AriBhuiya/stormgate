package routing_strategy

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"strings"
)

type SimpleRouting struct {
	Services *[]utils.ServiceConfig
}

func (r *SimpleRouting) Route(prefixPath *string) (*RouteEntry, error) {
	if prefixPath == nil || *prefixPath == "" {
		return nil, errors.New("invalid path")
	}

	var matched *utils.ServiceConfig
	var longest int

	for _, svc := range *r.Services {
		prefix := strings.TrimRight(svc.PathPrefix, "/")
		path := strings.TrimRight(*prefixPath, "/")

		if strings.HasPrefix(path, prefix) {
			if len(prefix) > longest {
				longest = len(prefix)
				matched = &svc
			}
		}
	}

	if matched != nil {
		return &RouteEntry{
			Path:    matched.PathPrefix,
			Service: matched,
		}, nil
	}

	return nil, errors.New("no matching service found")
}

func NewSimpleRouting(services *[]utils.ServiceConfig) *SimpleRouting {
	s := SimpleRouting{services}
	return &s
}
