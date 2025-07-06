package routing_strategy

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"reflect"
	"testing"
)

func TestHttpHybridRouting_Route(t *testing.T) {
	type fields struct {
		Depth1Map  map[string]*RouteEntry
		Depth2Map  map[string]*RouteEntry
		Depth3Map  map[string]*RouteEntry
		BaseRoute  *RouteEntry
		LongRoutes []*RouteEntry
	}
	type args struct {
		prefixPath *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RouteEntry
		wantErr bool
	}{
		{
			name: "match depth1 route",
			fields: fields{
				Depth1Map: map[string]*RouteEntry{
					"/auth": {
						Path: "/auth",
						Service: &utils.Service{
							Name:       "auth-service",
							PathPrefix: "/auth",
							Strategy:   "round-robin",
							Backends:   []string{"auth-1", "auth-2"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/auth/login")},
			want: &RouteEntry{
				Path: "/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/auth",
					Strategy:   "round-robin",
					Backends:   []string{"auth-1", "auth-2"},
				},
			},
			wantErr: false,
		},
		{
			name: "match depth2 route",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/api/v1": {
						Path: "/api/v1",
						Service: &utils.Service{
							Name:       "api-v1-service",
							PathPrefix: "/api/v1",
							Strategy:   "least-connections",
							Backends:   []string{"api-v1-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1/users")},
			want: &RouteEntry{
				Path: "/api/v1",
				Service: &utils.Service{
					Name:       "api-v1-service",
					PathPrefix: "/api/v1",
					Strategy:   "least-connections",
					Backends:   []string{"api-v1-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "match depth3 route",
			fields: fields{
				Depth3Map: map[string]*RouteEntry{
					"/api/v2/users": {
						Path: "/api/v2/users",
						Service: &utils.Service{
							Name:       "users-service",
							PathPrefix: "/api/v2/users",
							Strategy:   "ip-hash",
							Backends:   []string{"users-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v2/users/profile")},
			want: &RouteEntry{
				Path: "/api/v2/users",
				Service: &utils.Service{
					Name:       "users-service",
					PathPrefix: "/api/v2/users",
					Strategy:   "ip-hash",
					Backends:   []string{"users-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "match long route directly",
			fields: fields{
				LongRoutes: []*RouteEntry{
					{
						Path: "/api/v1/users/profile",
						Service: &utils.Service{
							Name:       "profile-service",
							PathPrefix: "/api/v1/users/profile",
							Strategy:   "random",
							Backends:   []string{"profile-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1/users/profile/myuser/something")},
			want: &RouteEntry{
				Path: "/api/v1/users/profile",
				Service: &utils.Service{
					Name:       "profile-service",
					PathPrefix: "/api/v1/users/profile",
					Strategy:   "random",
					Backends:   []string{"profile-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "fallback to base route",
			fields: fields{
				BaseRoute: &RouteEntry{
					Path: "/",
					Service: &utils.Service{
						Name:       "default-service",
						PathPrefix: "/",
						Strategy:   "round-robin",
						Backends:   []string{"default-1"},
					},
				},
			},
			args: args{prefixPath: ptr("/unknown/path")},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "default-service",
					PathPrefix: "/",
					Strategy:   "round-robin",
					Backends:   []string{"default-1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "no match, no base route",
			fields:  fields{},
			args:    args{prefixPath: ptr("/no/match")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "depth1 match with trailing slash",
			fields: fields{
				Depth1Map: map[string]*RouteEntry{
					"/auth": {
						Path: "/auth",
						Service: &utils.Service{
							Name:       "auth-service",
							PathPrefix: "/auth",
							Strategy:   "round-robin",
							Backends:   []string{"auth-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/auth/")},
			want: &RouteEntry{
				Path: "/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/auth",
					Strategy:   "round-robin",
					Backends:   []string{"auth-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "depth2 match with missing trailing slash",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/api/v1": {
						Path: "/api/v1",
						Service: &utils.Service{
							Name:       "v1-service",
							PathPrefix: "/api/v1",
							Strategy:   "ip-hash",
							Backends:   []string{"v1-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1")},
			want: &RouteEntry{
				Path: "/api/v1",
				Service: &utils.Service{
					Name:       "v1-service",
					PathPrefix: "/api/v1",
					Strategy:   "ip-hash",
					Backends:   []string{"v1-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "depth3 match with trailing slash and query string",
			fields: fields{
				Depth3Map: map[string]*RouteEntry{
					"/api/v1/products": {
						Path: "/api/v1/products",
						Service: &utils.Service{
							Name:       "product-service",
							PathPrefix: "/api/v1/products",
							Strategy:   "least-connections",
							Backends:   []string{"prod-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1/products/?limit=5")},
			want: &RouteEntry{
				Path: "/api/v1/products",
				Service: &utils.Service{
					Name:       "product-service",
					PathPrefix: "/api/v1/products",
					Strategy:   "least-connections",
					Backends:   []string{"prod-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "depth3 match with nested path beyond prefix",
			fields: fields{
				Depth3Map: map[string]*RouteEntry{
					"/api/v1/products": {
						Path: "/api/v1/products",
						Service: &utils.Service{
							Name:       "product-service",
							PathPrefix: "/api/v1/products",
							Strategy:   "round-robin",
							Backends:   []string{"prod-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1/products/123/reviews")},
			want: &RouteEntry{
				Path: "/api/v1/products",
				Service: &utils.Service{
					Name:       "product-service",
					PathPrefix: "/api/v1/products",
					Strategy:   "round-robin",
					Backends:   []string{"prod-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "reject path with double slashes",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/api/v2": {
						Path: "/api/v2",
						Service: &utils.Service{
							Name:       "v2-service",
							PathPrefix: "/api/v2",
							Strategy:   "random",
							Backends:   []string{"v2-1"},
						},
					},
				},
			},
			args:    args{prefixPath: ptr("/api//v2/users")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "match fallback base route",
			fields: fields{
				BaseRoute: &RouteEntry{
					Path: "/",
					Service: &utils.Service{
						Name:       "default-service",
						PathPrefix: "/",
						Strategy:   "round-robin",
						Backends:   []string{"default-1"},
					},
				},
			},
			args: args{prefixPath: ptr("/no/match/here")},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "default-service",
					PathPrefix: "/",
					Strategy:   "round-robin",
					Backends:   []string{"default-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "exact match on depth2 with query and trailing slash",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/files/v2": {
						Path: "/files/v2",
						Service: &utils.Service{
							Name:       "file-service",
							PathPrefix: "/files/v2",
							Strategy:   "round-robin",
							Backends:   []string{"fs-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/files/v2/?file=x.pdf")},
			want: &RouteEntry{
				Path: "/files/v2",
				Service: &utils.Service{
					Name:       "file-service",
					PathPrefix: "/files/v2",
					Strategy:   "round-robin",
					Backends:   []string{"fs-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "overlapping depth2 and depth3 routes, should pick depth3",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/api/v2": {
						Path: "/api/v2",
						Service: &utils.Service{
							Name:       "v2-service",
							PathPrefix: "/api/v2",
							Strategy:   "random",
							Backends:   []string{"v2"},
						},
					},
				},
				Depth3Map: map[string]*RouteEntry{
					"/api/v2/payments": {
						Path: "/api/v2/payments",
						Service: &utils.Service{
							Name:       "payment-service",
							PathPrefix: "/api/v2/payments",
							Strategy:   "round-robin",
							Backends:   []string{"pay1", "pay2"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v2/payments/history")},
			want: &RouteEntry{
				Path: "/api/v2/payments",
				Service: &utils.Service{
					Name:       "payment-service",
					PathPrefix: "/api/v2/payments",
					Strategy:   "round-robin",
					Backends:   []string{"pay1", "pay2"},
				},
			},
			wantErr: false,
		},
		{
			name: "shadowed prefix match should respect longest match",
			fields: fields{
				Depth1Map: map[string]*RouteEntry{
					"/shop": {
						Path: "/shop",
						Service: &utils.Service{
							Name:       "shop-service",
							PathPrefix: "/shop",
							Strategy:   "ip-hash",
							Backends:   []string{"shop-1"},
						},
					},
				},
				Depth2Map: map[string]*RouteEntry{
					"/shop/cart": {
						Path: "/shop/cart",
						Service: &utils.Service{
							Name:       "cart-service",
							PathPrefix: "/shop/cart",
							Strategy:   "round-robin",
							Backends:   []string{"cart-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/shop/cart/view")},
			want: &RouteEntry{
				Path: "/shop/cart",
				Service: &utils.Service{
					Name:       "cart-service",
					PathPrefix: "/shop/cart",
					Strategy:   "round-robin",
					Backends:   []string{"cart-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "case-sensitive mismatch — route not found",
			fields: fields{
				Depth1Map: map[string]*RouteEntry{
					"/admin": {
						Path: "/admin",
						Service: &utils.Service{
							Name:       "admin-service",
							PathPrefix: "/admin",
							Strategy:   "round-robin",
							Backends:   []string{"admin-1"},
						},
					},
				},
			},
			args:    args{prefixPath: ptr("/ADMIN")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "depth1 match should work with trailing slashes and query",
			fields: fields{
				Depth1Map: map[string]*RouteEntry{
					"/settings": {
						Path: "/settings",
						Service: &utils.Service{
							Name:       "settings-service",
							PathPrefix: "/settings",
							Strategy:   "round-robin",
							Backends:   []string{"s1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/settings/?theme=dark")},
			want: &RouteEntry{
				Path: "/settings",
				Service: &utils.Service{
					Name:       "settings-service",
					PathPrefix: "/settings",
					Strategy:   "round-robin",
					Backends:   []string{"s1"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid: path contains consecutive slashes",
			fields: fields{
				Depth2Map: map[string]*RouteEntry{
					"/api/v4": {
						Path: "/api/v4",
						Service: &utils.Service{
							Name:       "v4-service",
							PathPrefix: "/api/v4",
							Strategy:   "round-robin",
							Backends:   []string{"v4-1"},
						},
					},
				},
			},
			args:    args{prefixPath: ptr("/api//v4/data")},
			want:    nil,
			wantErr: true,
		},
		{
			name: "long route prefix match (not exact)",
			fields: fields{
				LongRoutes: []*RouteEntry{
					{
						Path: "/api/v3/logs/errors",
						Service: &utils.Service{
							Name:       "error-log-service",
							PathPrefix: "/api/v3/logs/errors",
							Strategy:   "least-connections",
							Backends:   []string{"error-log-1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v3/logs/errors/archive/2024")},
			want: &RouteEntry{
				Path: "/api/v3/logs/errors",
				Service: &utils.Service{
					Name:       "error-log-service",
					PathPrefix: "/api/v3/logs/errors",
					Strategy:   "least-connections",
					Backends:   []string{"error-log-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "no route match — fallback to base route",
			fields: fields{
				BaseRoute: &RouteEntry{
					Path: "/",
					Service: &utils.Service{
						Name:       "base-service",
						PathPrefix: "/",
						Strategy:   "round-robin",
						Backends:   []string{"base"},
					},
				},
			},
			args: args{prefixPath: ptr("/unknown/path/to/nothing")},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "base-service",
					PathPrefix: "/",
					Strategy:   "round-robin",
					Backends:   []string{"base"},
				},
			},
			wantErr: false,
		},
		{
			name: "prefix in Depth3Map but path has >3 segments — should use LongRoutes",
			fields: fields{
				Depth3Map: map[string]*RouteEntry{
					"/api/v1/users": {
						Path: "/api/v1/users",
						Service: &utils.Service{
							Name:       "users-service",
							PathPrefix: "/api/v1/users",
							Strategy:   "round-robin",
							Backends:   []string{"user1"},
						},
					},
				},
				LongRoutes: []*RouteEntry{
					{
						Path: "/api/v1/users/admin",
						Service: &utils.Service{
							Name:       "admin-service",
							PathPrefix: "/api/v1/users/admin",
							Strategy:   "ip-hash",
							Backends:   []string{"admin1"},
						},
					},
				},
			},
			args: args{prefixPath: ptr("/api/v1/users/admin/logs")},
			want: &RouteEntry{
				Path: "/api/v1/users/admin",
				Service: &utils.Service{
					Name:       "admin-service",
					PathPrefix: "/api/v1/users/admin",
					Strategy:   "ip-hash",
					Backends:   []string{"admin1"},
				},
			},
			wantErr: false,
		},
		{
			name: "case-sensitive mismatch in long route",
			fields: fields{
				LongRoutes: []*RouteEntry{
					{
						Path: "/Metrics/Prometheus",
						Service: &utils.Service{
							Name:       "case-sensitive-service",
							PathPrefix: "/Metrics/Prometheus",
							Strategy:   "round-robin",
							Backends:   []string{"prom-2"},
						},
					},
				},
			},
			args:    args{prefixPath: ptr("/metrics/prometheus/")},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpHybridRouting{
				Depth1Map:  tt.fields.Depth1Map,
				Depth2Map:  tt.fields.Depth2Map,
				Depth3Map:  tt.fields.Depth3Map,
				BaseRoute:  tt.fields.BaseRoute,
				LongRoutes: tt.fields.LongRoutes,
			}
			got, err := r.Route(tt.args.prefixPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Route() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func ptr(s string) *string {
	return &s
}
