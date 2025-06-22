package routers

import (
	"fmt"
	"github.com/aribhuiya/stormgate/internal/routing_strategy"
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
	"net/http"
	"time"
)

type HttpRouter struct {
	config          Config
	routingStrategy *routing_strategy.HttpHybridRouting
	//routingStrategy *routing_strategy.SimpleRouting
}

type Config struct {
	BindIp         string
	BindPort       int32
	MaxConnections int32
	ReadTimeOutMs  int64
	WriteTimeOutMs int64
}

// NewHttpRouter Use to Create a HttpRouter Instance from a HttpRouter.Config object with appropriate defaults
func NewHttpRouter(cfg Config, serviceConfigs *[]utils.ServiceConfig) *HttpRouter {
	if cfg.BindIp == "" {
		cfg.BindIp = "127.0.0.1"
	}
	if cfg.BindPort == 0 {
		cfg.BindPort = 10000
	}
	return &HttpRouter{
		config:          cfg,
		routingStrategy: routing_strategy.NewHttpHybridRouting(serviceConfigs),
		//routingStrategy: routing_strategy.NewSimpleRouting(serviceConfigs),
	}
}

// NewRouterFromConfig Use to Directly get a Router instance from the yaml config with appropriate defaults
func NewRouterFromConfig(config utils.Config) Router {
	serverConfig := &config.Server
	cfg := Config{
		BindIp:         serverConfig.BindIp,
		BindPort:       serverConfig.BindPort,
		ReadTimeOutMs:  serverConfig.ReadTimeOut,
		WriteTimeOutMs: serverConfig.WriteTimeOut,
	}
	return NewHttpRouter(cfg, &config.Services)
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
