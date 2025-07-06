package balancers

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestRoundRobin_PickBackend(t *testing.T) {
	type fields struct {
		counter atomic.Uint64
		service *utils.Service
		n       uint64
	}
	type args struct {
		in0 *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "First backend selected",
			fields: fields{
				counter: atomic.Uint64{},
				service: &utils.Service{
					Backends: []string{"http://localhost:9001", "http://localhost:9002"},
				},
				n: 2,
			},
			args:    args{in0: &http.Request{}},
			want:    "http://localhost:9002",
			wantErr: false,
		},
		{
			name: "Wrap around to first backend",
			fields: fields{
				counter: func() atomic.Uint64 {
					var c atomic.Uint64
					c.Store(1)
					return c
				}(),
				service: &utils.Service{
					Backends: []string{"http://localhost:9001", "http://localhost:9002"},
				},
				n: 2,
			},
			args:    args{in0: &http.Request{}},
			want:    "http://localhost:9001", // counter goes to 2 → index = 0
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RoundRobin{
				counter: tt.fields.counter,
				service: tt.fields.service,
				n:       tt.fields.n,
			}
			got, err := r.PickBackend(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("PickBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PickBackend() got = %v, want %v", got, tt.want)
			}
		})
	}
}
