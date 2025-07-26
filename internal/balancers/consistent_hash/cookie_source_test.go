package consistent_hash

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieSource_getSource(t *testing.T) {
	tests := []struct {
		name            string
		req             *http.Request
		source          *cookie_source
		expectInject    bool
		expectNonEmpty  bool
		expectedContext bool
	}{
		{
			name: "cookie present with value and key",
			req: func() *http.Request {
				val := base64.StdEncoding.EncodeToString([]byte(`{"X-User-ID":"abc123"}`))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: val,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: true,
		},
		{
			name: "cookie missing and inject true",
			req: func() *http.Request {
				return httptest.NewRequest("GET", "/api/v2", nil)
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: true,
			},
			expectNonEmpty:  true,
			expectInject:    true,
			expectedContext: true,
		},
		{
			name: "cookie present but missing key, fallback inject false",
			req: func() *http.Request {
				val := base64.StdEncoding.EncodeToString([]byte(`{"OtherKey":"val"}`))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: val,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: false,
		},
		{
			name: "cookie present but plain string, key is empty",
			req: func() *http.Request {
				val := base64.StdEncoding.EncodeToString([]byte(`some_value`))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: val,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "",
				injectIfMissing: false,
			},
			expectNonEmpty: true,
		},
		{
			name: "cookie with malformed base64",
			req: func() *http.Request {
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: "!!!invalid-base64",
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: false,
		},
		{
			name: "cookie with base64 but invalid JSON",
			req: func() *http.Request {
				invalidJson := base64.StdEncoding.EncodeToString([]byte("{ not json }"))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: invalidJson,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: false,
		},
		{
			name: "cookie with JSON number value for key",
			req: func() *http.Request {
				jsonValue := base64.StdEncoding.EncodeToString([]byte(`{"X-User-ID":12345}`))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: jsonValue,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: true, // because it will be coerced via fmt.Sprintf
		},
		{
			name: "cookie with null value for key",
			req: func() *http.Request {
				jsonValue := base64.StdEncoding.EncodeToString([]byte(`{"X-User-ID":null}`))
				r := httptest.NewRequest("GET", "/api/v2", nil)
				r.AddCookie(&http.Cookie{
					Name:  "cookie_name",
					Value: jsonValue,
				})
				return r
			}(),
			source: &cookie_source{
				cookieName:      "cookie_name",
				cookieKey:       "X-User-ID",
				injectIfMissing: false,
			},
			expectNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.source.getSource(tt.req)

			if tt.expectNonEmpty && val == "" {
				t.Errorf("expected non-empty string, got empty")
			}
			if !tt.expectNonEmpty && val != "" {
				t.Errorf("expected empty string, got '%s'", val)
			}

			if tt.expectedContext {
				cVal := tt.req.Context().Value("inject_cookie")
				if cVal == nil {
					t.Errorf("expected context value 'inject_cookie' to be set")
				}
			}
		})
	}
}
