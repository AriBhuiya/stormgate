package consistent_hash

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aribhuiya/stormgate/internal/utils"
	"github.com/google/uuid"
	"net/http"
)

type cookie_source struct {
	cookieName      string
	cookieKey       string
	injectIfMissing bool
}

func (c *cookie_source) getSource(req *http.Request) string {
	cookie, err := req.Cookie(c.cookieName)
	if err != nil {
		if c.injectIfMissing {
			return generateUUID()
		}
		return ""
	}

	if c.cookieKey == "" {
		return cookie.Value
	}

	decoded, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return ""
	}

	var parsed map[string]any
	err = json.Unmarshal(decoded, &parsed)
	if err != nil {
		return ""
	}
	val, exists := parsed[c.cookieKey]
	if !exists {
		return ""
	}
	strVal, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return strVal
}

func generateUUID() string {
	return uuid.New().String()
}

func NewCookieSource(service *utils.Service) (*cookie_source, error) {
	name, ok := service.StrategyConfig["name"].(string)
	if !ok {
		return nil, errors.New("cookie name is required")
	}
	key, ok := service.StrategyConfig["key"].(string)
	if !ok {
		key = ""
	}

	injectIfMissing := false
	injectIfMissingRaw, exists := service.StrategyConfig["inject_if_missing"]
	if exists {
		val, ok := injectIfMissingRaw.(bool)
		if !ok {
			return nil, errors.New("inject_if_missing value must be a true/false")
		}
		injectIfMissing = val
	}

	return &cookie_source{
		cookieName:      name,
		cookieKey:       key,
		injectIfMissing: injectIfMissing,
	}, nil
}
