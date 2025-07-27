package health_checker

import "github.com/aribhuiya/stormgate/internal/stormgate"

type HealthChecker interface {
	CheckHealth(service stormgate.Service)
}
