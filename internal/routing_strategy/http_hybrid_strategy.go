package routing_strategy

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"sort"
	"strings"
)

type HttpHybridRouting struct {
	Depth1Map  map[string]*RouteEntry
	Depth2Map  map[string]*RouteEntry
	Depth3Map  map[string]*RouteEntry
	BaseRoute  *RouteEntry
	LongRoutes []*RouteEntry
}

type RouteEntry struct {
	Path    string
	Service *utils.ServiceConfig
}

func NewHttpHybridRouting(services *[]utils.ServiceConfig) *HttpHybridRouting {
	r := HttpHybridRouting{
		Depth1Map:  make(map[string]*RouteEntry),
		Depth2Map:  make(map[string]*RouteEntry),
		Depth3Map:  make(map[string]*RouteEntry),
		LongRoutes: []*RouteEntry{},
	}
	for _, service := range *services {
		normalisedPath := utils.NormalizePath(service.PathPrefix)
		segments := strings.Split(strings.TrimPrefix(normalisedPath, "/"), "/")
		switch len(segments) {
		case 1:
			if normalisedPath == "/" { //base
				r.BaseRoute = &RouteEntry{Path: normalisedPath, Service: &service}
				break
			}
			r.Depth1Map[normalisedPath] = &RouteEntry{Path: normalisedPath, Service: &service}
		case 2:
			r.Depth2Map[normalisedPath] = &RouteEntry{Path: normalisedPath, Service: &service}
		case 3:
			r.Depth3Map[normalisedPath] = &RouteEntry{Path: normalisedPath, Service: &service}
		default:
			r.LongRoutes = append(r.LongRoutes, &RouteEntry{Path: normalisedPath, Service: &service})
		}
	}
	sort.SliceStable(r.LongRoutes, func(i, j int) bool {
		return len(r.LongRoutes[i].Path) > len(r.LongRoutes[j].Path)
	})
	return &r
}

func (r *HttpHybridRouting) Route(prefixPath *string) (*RouteEntry, error) {
	depth3, depth2, depth1, isMoreThanThree := utils.ExtractPrefixes(*prefixPath)

	if isMoreThanThree {
		for _, route := range r.LongRoutes {
			prefix := route.Path
			if strings.HasPrefix(*prefixPath, prefix) {
				if len(*prefixPath) == len(prefix) || (*prefixPath)[len(prefix)] == '/' || (*prefixPath)[len(prefix)] == '?' {
					return route, nil
				}
			}
		}
	}

	if depth3 != "" {
		if route, ok := r.Depth3Map[depth3]; ok {
			return route, nil
		}
	}
	if depth2 != "" {
		if route, ok := r.Depth2Map[depth2]; ok {
			return route, nil
		}
	}
	if depth1 != "" {
		if route, ok := r.Depth1Map[depth1]; ok {
			return route, nil
		}
	}

	// If base path is defined
	if r.BaseRoute != nil {
		return r.BaseRoute, nil
	}

	return nil, errors.New("no matching route found")
}
