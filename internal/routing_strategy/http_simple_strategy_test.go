package routing_strategy

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"reflect"
	"testing"
)

func TestSimpleRouting_Route(t *testing.T) {
	type fields struct {
		Services *[]utils.Service
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
			name: "exact prefix match",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth-service", PathPrefix: "/api/auth"},
					{Name: "user-service", PathPrefix: "/api/user"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/api/auth",
					Strategy:   "",
					Backends:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "longest prefix match",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "api-root", PathPrefix: "/api"},
					{Name: "auth-service", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth/login"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/api/auth",
					Strategy:   "",
					Backends:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "no matching prefix",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "admin-service", PathPrefix: "/admin"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "prefix overlaps but not a true match",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth-service", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth123"), // shouldn't match
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil prefix path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "api", PathPrefix: "/api"},
				},
			},
			args: args{
				prefixPath: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "match with trailing slash",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth-service", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth/"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/api/auth",
				},
			},
			wantErr: false,
		},
		{
			name: "match with query param",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth-service", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth?token=abc"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth-service",
					PathPrefix: "/api/auth",
				},
			},
			wantErr: false,
		},
		{
			name: "deep nested match - longest prefix wins",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
					{Name: "api", PathPrefix: "/api"},
					{Name: "auth", PathPrefix: "/api/auth"},
					{Name: "login", PathPrefix: "/api/auth/login"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth/login/otp"),
			},
			want: &RouteEntry{
				Path: "/api/auth/login",
				Service: &utils.Service{
					Name:       "login",
					PathPrefix: "/api/auth/login",
				},
			},
			wantErr: false,
		},
		{
			name: "match root fallback",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
				},
			},
			args: args{
				prefixPath: StringPtr("/unknown/path"),
			},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "root",
					PathPrefix: "/",
				},
			},
			wantErr: false,
		},
		{
			name: "disallow double slash in path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth-service", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api//auth"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "match prefix before query param with slash",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "user-service", PathPrefix: "/api/user"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/user?id=1"),
			},
			want: &RouteEntry{
				Path: "/api/user",
				Service: &utils.Service{
					Name:       "user-service",
					PathPrefix: "/api/user",
				},
			},
			wantErr: false,
		},
		{
			name: "partial overlap should not match",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "user-service", PathPrefix: "/api/user"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/userdata"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "exact match with single slash root path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
				},
			},
			args: args{
				prefixPath: StringPtr("/"),
			},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "root",
					PathPrefix: "/",
				},
			},
			wantErr: false,
		},
		{
			name: "longer prefix match over shorter prefix",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "api", PathPrefix: "/api"},
					{Name: "auth", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth/profile"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth",
					PathPrefix: "/api/auth",
				},
			},
			wantErr: false,
		},
		{
			name: "no match for partial prefix overlap",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/authentication"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "match with trailing slash in request path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/auth/"),
			},
			want: &RouteEntry{
				Path: "/api/auth",
				Service: &utils.Service{
					Name:       "auth",
					PathPrefix: "/api/auth",
				},
			},
			wantErr: false,
		},
		{
			name: "match with query params in request",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "user", PathPrefix: "/api/user"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api/user?id=42"),
			},
			want: &RouteEntry{
				Path: "/api/user",
				Service: &utils.Service{
					Name:       "user",
					PathPrefix: "/api/user",
				},
			},
			wantErr: false,
		},
		{
			name: "no match with double slash in path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "auth", PathPrefix: "/api/auth"},
				},
			},
			args: args{
				prefixPath: StringPtr("/api//auth"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "match fallback to root",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
				},
			},
			args: args{
				prefixPath: StringPtr("/anything/unmatched"),
			},
			want: &RouteEntry{
				Path: "/",
				Service: &utils.Service{
					Name:       "root",
					PathPrefix: "/",
				},
			},
			wantErr: false,
		},
		{
			name: "empty string as path",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
				},
			},
			args: args{
				prefixPath: StringPtr(""),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil path input",
			fields: fields{
				Services: &[]utils.Service{
					{Name: "root", PathPrefix: "/"},
				},
			},
			args: args{
				prefixPath: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SimpleRouting{
				Services: tt.fields.Services,
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

func StringPtr(s string) *string {
	return &s
}
