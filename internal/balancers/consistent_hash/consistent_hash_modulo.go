package consistent_hash

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"github.com/cespare/xxhash/v2"
	"net/http"
	"strings"
)

type HashModulo struct {
	service            *utils.Service
	source             hashSource
	fallbackToIpSource *ipSource
}

type hashSource interface {
	getSource(req *http.Request) string
}

const (
	SOURCE_IP     = "IP"
	SOURCE_HEADER = "HEADER"
	SOURCE_COOKIE = "COOKIE"
)

func NewHashModulo(service *utils.Service) (*HashModulo, error) {
	if len(service.Backends) == 0 {
		err := errors.New("no available backends")
		return nil, err
	}
	source, ok := service.StrategyConfig["source"].(string)
	if !ok {
		return nil, errors.New("source not defined for consistent_hash")
	}
	var hashKeySource hashSource
	var err error
	switch strings.ToUpper(source) {
	case SOURCE_IP:
		hashKeySource = NewIPSource(service)
	case SOURCE_HEADER:
		hashKeySource, err = NewHeaderSource(service)
		if err != nil {
			return nil, err
		}
	case SOURCE_COOKIE:
		hashKeySource, err = NewCookieSource(service)
		if err != nil {
			return nil, err
		}
	}

	var fallbackToIP *ipSource = nil
	fallbackToIPRaw, exists := service.StrategyConfig["fallback_to_ip"]
	if exists {
		val, ok := fallbackToIPRaw.(bool)
		if !ok {
			return nil, errors.New("fallback_to_ip value must be a true/false")
		}
		if val {
			fallbackToIP = NewIPSource(service)
		}
	}

	return &HashModulo{
		service:            service,
		source:             hashKeySource,
		fallbackToIpSource: fallbackToIP,
	}, nil
}

func hashString(s string) uint64 {
	return xxhash.Sum64String(s)
}

func (h *HashModulo) PickBackend(req *http.Request) (string, error) {
	if len(h.service.Backends) == 0 {
		return "", errors.New("no healthy backends")
	}
	key := h.source.getSource(req)

	if key == "" && h.fallbackToIpSource != nil {
		key = h.fallbackToIpSource.getSource(req)
	}

	if key == "" {
		return "", errors.New("unable to derive key for hashing")
	}

	hash := hashString(key)
	index := int(hash % uint64(len(h.service.Backends)))
	return h.service.Backends[index], nil
}

func (h *HashModulo) SetHealthyBackends(healthyBackends []string) {
	if hasChanged := utils.HasBackendChanged(h.service.Backends, healthyBackends); !hasChanged {
		return
	}
	h.service.Backends = healthyBackends
}
