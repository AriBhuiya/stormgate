package consistent_hash

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"net"
	"net/http"
)

type ipSource struct {
}

func NewIPSource(_ *utils.Service) *ipSource {
	return &ipSource{}
}

func (s *ipSource) getSource(req *http.Request) string {
	// Check for common headers first (e.g. when behind a reverse proxy)
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.Header.Get("X-Real-IP")
	}

	if ip == "" {
		ip, _, _ = net.SplitHostPort(req.RemoteAddr)
	}

	return ip
}
