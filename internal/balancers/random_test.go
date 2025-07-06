package balancers

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"testing"
)

func TestRandom_PickBackend(t *testing.T) {
	type fields struct {
		service utils.Service
		seed    int32
	}
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Select random backend from 2 backends",
			fields: fields{
				service: utils.Service{
					Backends: []string{"http://localhost:9001", "http://localhost:9002"},
				},
				seed: 42,
			},
			args:    args{request: &http.Request{}},
			want:    "", // we won't check exact match
			wantErr: false,
		},
		{
			name: "Select random backend from single backend",
			fields: fields{
				service: utils.Service{
					Backends: []string{"http://localhost:9001"},
				},
				seed: 7,
			},
			args:    args{request: &http.Request{}},
			want:    "", // will validate against Backends
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Random{
				service: tt.fields.service,
				seed:    tt.fields.seed,
			}
			got, err := r.PickBackend(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("PickBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			found := false
			for _, b := range tt.fields.service.Backends {
				if got == b {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("PickBackend() got = %v, which is not a valid backend", got)
			}
		})
	}
}
