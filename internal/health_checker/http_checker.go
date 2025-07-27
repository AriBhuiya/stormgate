package health_checker

import (
	"github.com/aribhuiya/stormgate/internal/stormgate"
	"io"
	"net/http"
	"strings"
)

type HttpChecker struct {
	IntervalMs uint64
	EndPoint   string
	Service    stormgate.Service
}

func NewHttpChecker(service stormgate.Service, endPoint string, interval uint64) *HttpChecker {
	return &HttpChecker{
		interval,
		endPoint,
		service,
	}
}

func (h HttpChecker) CheckHealth(service stormgate.Service) {
	backends := service.Config.Backends
	for _, backend := range backends {
		endpoint := strings.TrimRight(backend, "/") + "/" + h.EndPoint
		if isHealthy := isHealthy(endpoint); !isHealthy {
			//TODO: remove the endpoint
		}
	}
}

func isHealthy(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return true
	}
	return false
}
