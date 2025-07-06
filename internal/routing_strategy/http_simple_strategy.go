package routing_strategy

import (
	"errors"
	"fmt"
	"github.com/aribhuiya/stormgate/internal/utils"
	"strings"
)

type SimpleRouting struct {
	Services *[]utils.Service
}

func (r *SimpleRouting) Route(prefixPath *string) (*RouteEntry, error) {
	if prefixPath == nil || *prefixPath == "" {
		return nil, errors.New("invalid path")
	}

	var matched *utils.Service
	var longest int

	for _, svc := range *r.Services {
		prefix := svc.PathPrefix
		if prefix != "/" {
			prefix = strings.TrimRight(prefix, "/")
		}

		path := *prefixPath
		if path != "/" {
			path = strings.TrimRight(path, "/")
		}

		if strings.HasPrefix(path, prefix) {
			if prefix == "/" || len(path) == len(prefix) || path[len(prefix)] == '/' || path[len(prefix)] == '?' {
				if len(prefix) > longest {
					longest = len(prefix)
					matched = &svc
				}
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

func NewSimpleRouting(services *[]utils.Service) *SimpleRouting {
	fmt.Println("Simple Routing used ...")
	s := SimpleRouting{services}
	return &s
}
