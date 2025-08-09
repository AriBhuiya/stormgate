package health_checker

import "time"

type HealthChecker interface {
	CheckHealth() []string
	CheckAndUpdateBalancer()
	GetInterval() time.Duration
}
