package consistent_hash

import (
	"errors"
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
)

type headerSource struct {
	keyName string
}

func NewHeaderSource(service *utils.Service) (*headerSource, error) {
	key, ok := service.StrategyConfig["key"].(string)
	if !ok {
		return nil, errors.New("key is required for source header")
	}
	return &headerSource{
		keyName: key,
	}, nil
}

func (c *headerSource) getSource(req *http.Request) string {
	return req.Header.Get(c.keyName)
}
