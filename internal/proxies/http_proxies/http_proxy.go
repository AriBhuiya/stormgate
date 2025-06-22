package http_proxies

import (
	"net/http"
)

type Proxy interface {
	Forward(w http.ResponseWriter, req *http.Request, forwardingEndpoint *string)
}
