package stormgate

import (
	"fmt"
	"github.com/aribhuiya/stormgate/internal/balancers"
	"github.com/aribhuiya/stormgate/internal/proxies/http_proxies"
	"github.com/aribhuiya/stormgate/internal/routing_strategy"
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
	"net/http"
	"time"
)

type StormGate struct {
	ServerConfig ServerConfig
	Services     map[string]*service
	routing_strategy.RoutingStrategy
	Proxy http_proxies.Proxy
}

type ServerConfig struct {
	BindIp         string
	BindPort       int32
	MaxConnections int32
	ReadTimeOutMs  int64
	WriteTimeOutMs int64
}

func NewStormGate(config utils.Config) (*StormGate, error) {
	serverConfig := &config.Server
	cfg := ServerConfig{
		BindIp:         serverConfig.BindIp,
		BindPort:       serverConfig.BindPort,
		ReadTimeOutMs:  serverConfig.ReadTimeOut,
		WriteTimeOutMs: serverConfig.WriteTimeOut,
	}
	services, err := BuildServicesFromConfig(config.Services)
	if err != nil {
		return nil, err
	}
	return &StormGate{
		ServerConfig:    cfg,
		Services:        services,
		RoutingStrategy: routing_strategy.CreateRoutingStrategy(config.Balancer.RoutingStrategy, &config.Services),
		Proxy:           http_proxies.NewBasicProxy(),
	}, nil
}

func BuildServicesFromConfig(services []utils.Service) (map[string]*service, error) {
	servicesMap := make(map[string]*service)
	for _, svcCfg := range services {
		balancer, err := balancers.Create(svcCfg.Strategy, &svcCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create balancer for service %s: %w", svcCfg.Name, err)
		}
		svc := &service{
			config:   svcCfg,
			balancer: balancer,
		}
		servicesMap[svcCfg.PathPrefix] = svc
	}
	return servicesMap, nil
}

func (s *StormGate) Serve() {
	addr := fmt.Sprintf("%s:%d", s.ServerConfig.BindIp, s.ServerConfig.BindPort)
	err := http.ListenAndServe(addr, s)
	if err != nil {
		log.Fatalln("ListenAndServe: Can't bind to port. Make sure the port is available", err)
	}
}

// implicitly implements HTTP Serve
func (s *StormGate) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find routing prefix
	route, err := s.RoutingStrategy.Route(&req.URL.Path)

	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Find Service
	service := s.Services[route.Service.PathPrefix]
	if service == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Use Balancer
	forwardPath, err := service.balancer.PickBackend(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("E-1 Internal Server Error %s", err), http.StatusInternalServerError)
		return
	}

	val := req.Context().Value("inject_cookie")
	if cookieVal, ok := val.(string); ok {
		path := service.config.PathPrefix
		println("Setting cookie to " + cookieVal)
		http.SetCookie(w, &http.Cookie{
			Name:     "stormgate-id",
			Value:    cookieVal,
			Path:     path,
			HttpOnly: true,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
		})
	}

	s.Proxy.Forward(w, req, &forwardPath)

}
