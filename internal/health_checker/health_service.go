package health_checker

import (
	"context"
	"fmt"
	"github.com/aribhuiya/stormgate/internal/stormgate"
	"strings"
	"time"
)

type HealthCheckerService struct {
	checkers []HealthChecker
	cancel   context.CancelFunc
}

func NewHealthCheckerService(services map[string]*stormgate.Service) *HealthCheckerService {
	var checkers []HealthChecker

	for _, svc := range services {
		if svc.Config.Health == nil {
			continue // skip services without health config
		}

		healthCfg := svc.Config.Health

		switch strings.ToLower(healthCfg.Type) {
		case "http":
			if healthCfg.Endpoint == "" {
				panic(fmt.Sprintf("Health config error in service '%s': missing 'health-endpoint'", svc.Config.Name))
			}
			if healthCfg.Frequency <= 0 {
				panic(fmt.Sprintf("Health config error in service '%s': 'frequency' must be greater than 0", svc.Config.Name))
			}
			httpChecker := NewHttpChecker(*svc, healthCfg.Endpoint, uint64(healthCfg.Frequency))
			checkers = append(checkers, httpChecker)
		default:
			panic(fmt.Sprintf("Health config error in service '%s': unsupported health type '%s'", svc.Config.Name, healthCfg.Type))
		}
	}

	return &HealthCheckerService{
		checkers: checkers,
	}
}

func (h *HealthCheckerService) StartService() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	for _, checker := range h.checkers {
		go func(c HealthChecker) {
			ticker := time.NewTicker(c.GetInterval())
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					c.CheckAndUpdateBalancer()
				}
			}
		}(checker)
	}
}

func (h *HealthCheckerService) StopService() {
	if h.cancel != nil {
		h.cancel()
	}
}
