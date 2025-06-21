package routers

import (
	"fmt"
	"github.com/aribhuiya/stormgate/internal/utils"
	"log"
	"net/http"
)

type HttpRouter struct {
	config Config
}

type Config struct {
	BindIp         string
	BindPort       int32
	MaxConnections int32
	ReadTimeOutMs  int64
	WriteTimeOutMs int64
}

// NewHttpRouter Use to Create a HttpRouter Instance from a HttpRouter.Config object with appropriate defaults
func NewHttpRouter(cfg Config) *HttpRouter {
	if cfg.BindIp == "" {
		cfg.BindIp = "127.0.0.1"
	}
	if cfg.BindPort == 0 {
		cfg.BindPort = 10000
	}
	return &HttpRouter{
		config: cfg,
	}
}

// NewRouterFromConfig Use to Directly get a Router instance from the yaml config with appropriate defaults
func NewRouterFromConfig(serverconfig utils.ServerConfig) Router {
	cfg := Config{
		BindIp:         serverconfig.BindIp,
		BindPort:       serverconfig.BindPort,
		ReadTimeOutMs:  serverconfig.ReadTimeOut,
		WriteTimeOutMs: serverconfig.WriteTimeOut,
	}
	return NewHttpRouter(cfg)
}

func (r *HttpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Got request")
	fmt.Println(req.URL.Path)
}

func (r *HttpRouter) Serve() {
	addr := fmt.Sprintf("%s:%d", r.config.BindIp, r.config.BindPort)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalln("ListenAndServe: Can't bind to port. Make sure the port is available", err)
	}
}
