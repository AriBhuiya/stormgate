package consistent_hash

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"testing"
)

func TestHashModulo_PickBackend(t *testing.T) {
	tests := []struct {
		name          string
		service       *utils.Service
		request       *http.Request
		config        map[string]any
		expectError   bool
		expectBackend bool
	}{
		{
			name: "IP-based hashing",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "ip",
				},
			},
			request:       &http.Request{RemoteAddr: "192.168.1.1:12345"},
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Header-based hashing with valid header",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "header",
					"key":    "X-User-ID",
				},
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.Header.Set("X-User-ID", "user123")
				return req
			}(),
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Header-based missing header with fallback to IP",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source":         "header",
					"key":            "X-User-ID",
					"fallback_to_ip": true,
				},
			},
			request:       &http.Request{RemoteAddr: "10.0.0.1:4567"},
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Missing key and no fallback",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "header",
					"key":    "missing",
				},
			},
			request:       &http.Request{},
			expectError:   true,
			expectBackend: false,
		},
		{
			name: "Cookie-based hashing - plain value",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "cookie",
					"name":   "session_id",
				},
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
				return req
			}(),
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Cookie-based hashing - JSON with key",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source":     "cookie",
					"name":       "user",
					"cookie_key": "id",
				},
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{Name: "user", Value: `{"id":"xyz"}`})
				return req
			}(),
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Cookie missing - fallback to IP",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source":         "cookie",
					"name":           "missing",
					"fallback_to_ip": true,
				},
			},
			request:       &http.Request{RemoteAddr: "127.0.0.1:1234"},
			expectError:   false,
			expectBackend: true,
		},
		{
			name: "Cookie missing - no fallback",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "cookie",
					"name":   "missing",
				},
			},
			request:       &http.Request{},
			expectError:   true,
			expectBackend: false,
		},
		{
			name: "Invalid source type",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": 123,
				},
			},
			request:       &http.Request{},
			expectError:   true,
			expectBackend: false,
		},
		{
			name: "Missing source config",
			service: &utils.Service{
				Backends:       []string{"A", "B"},
				StrategyConfig: map[string]any{},
			},
			request:       &http.Request{},
			expectError:   true,
			expectBackend: false,
		},
		{
			name: "Header based hashing",
			service: &utils.Service{
				Backends: []string{"A", "B"},
				StrategyConfig: map[string]any{
					"source": "header",
					"key":    "X-User-ID",
				},
			},
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.Header.Set("X-User-ID", "user42")
				return req
			}(),
			expectError:   false,
			expectBackend: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewHashModulo(tt.service)
			if tt.expectError {
				if err != nil {
					return
				}
			}
			backend, err := h.PickBackend(tt.request)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if (err != nil) != tt.expectError {
				t.Errorf("PickBackend() error = %v, expected error = %v", err, tt.expectError)
			}
			if tt.expectBackend && backend == "" {
				t.Errorf("Expected a backend, got empty string")
			}

		})
	}
}
