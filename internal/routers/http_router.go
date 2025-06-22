package routers

import (
	"fmt"
	"github.com/aribhuiya/stormgate/internal/proxies/http_proxies"
	"github.com/aribhuiya/stormgate/internal/routing_strategy"
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
	"net/http"
	"time"
)

type HttpRouter struct {
	config          ServerConfig
	routingStrategy routing_strategy.RoutingStrategy
	proxy           http_proxies.Proxy
}

type ServerConfig struct {
	BindIp         string
	BindPort       int32
	MaxConnections int32
	ReadTimeOutMs  int64
	WriteTimeOutMs int64
}

// NewHttpRouter Use to Create a HttpRouter Instance from a HttpRouter.ServerConfig object with appropriate defaults
func NewHttpRouter(serverCfg ServerConfig, serviceConfigs *[]utils.Services, strategyName string) *HttpRouter {
	if serverCfg.BindIp == "" {
		serverCfg.BindIp = "127.0.0.1"
	}
	if serverCfg.BindPort == 0 {
		serverCfg.BindPort = 10000
	}
	routingStrategy := routing_strategy.CreateRoutingStrategy(strategyName, serviceConfigs)
	return &HttpRouter{
		config:          serverCfg,
		routingStrategy: routingStrategy,
		proxy:           http_proxies.NewBasicProxy(),
	}
}

// NewRouterFromConfig Use to Directly get a Router instance from the yaml config with appropriate defaults
func NewRouterFromConfig(config utils.Config) Router {
	serverConfig := &config.Server
	cfg := ServerConfig{
		BindIp:         serverConfig.BindIp,
		BindPort:       serverConfig.BindPort,
		ReadTimeOutMs:  serverConfig.ReadTimeOut,
		WriteTimeOutMs: serverConfig.WriteTimeOut,
	}
	return NewHttpRouter(cfg, &config.Services, config.Balancer.RoutingStrategy)
}

func (r *HttpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//fmt.Println("Got request")
	//fmt.Println(req.URL.Path)

	start := time.Now()
	route, err := r.routingStrategy.Route(&req.URL.Path)
	routingDuration := time.Since(start)

	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// TODO: Call goes to Balancer when ready
	r.proxy.Forward(w, req, &route.Service.Backends[0])

	_, err = fmt.Fprintf(w, `{
  		"matched_path": "%s",
  		"service": "%s",
  		"duration_ns": %d
	}`, route.Path, route.Service.Name, routingDuration.Nanoseconds())
}

func (r *HttpRouter) Serve() {
	addr := fmt.Sprintf("%s:%d", r.config.BindIp, r.config.BindPort)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalln("ListenAndServe: Can't bind to port. Make sure the port is available", err)
	}
}
