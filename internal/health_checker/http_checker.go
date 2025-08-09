package health_checker

import (
	"github.com/aribhuiya/stormgate/internal/stormgate"
	"io"
	"net/http"
	"strings"
	"time"
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

func (h *HttpChecker) CheckHealth() []string {
	backends := h.Service.Config.Backends
	var healthyBackends []string

	for _, backend := range backends {
		endpoint := strings.TrimRight(backend, "/") + "/" + h.EndPoint
		if isHealthy(endpoint) {
			healthyBackends = append(healthyBackends, backend)
		}
	}

	return healthyBackends
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

func (h *HttpChecker) CheckAndUpdateBalancer() {
	healthyBackends := h.CheckHealth()
	h.Service.Balancer.SetHealthyBackends(healthyBackends)
}

func (h *HttpChecker) GetInterval() time.Duration {
	return time.Duration(h.IntervalMs) * time.Millisecond
}
