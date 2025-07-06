package balancers

import (
	"github.com/aribhuiya/stormgate/internal/utils"
	"net/http"
	"testing"
)

func TestRoundRobin_PickBackend_MultipleCalls(t *testing.T) {
	backends := []string{"http://localhost:9001", "http://localhost:9002", "http://localhost:9003"}
	service := &utils.Service{Backends: backends}
	rr, err := NewRoundRobin(service)
	if err != nil {
		t.Fatalf("unexpected error creating RoundRobin: %v", err)
	}

	request := &http.Request{}
	expectedSequence := []string{
		"http://localhost:9001", // 0
		"http://localhost:9002", // 1
		"http://localhost:9003", // 2
		"http://localhost:9001", // 3 % 3 = 0
		"http://localhost:9002", // 4 % 3 = 1
		"http://localhost:9003", // 5 % 3 = 2
	}

	for i, expected := range expectedSequence {
		got, err := rr.PickBackend(request)
		if err != nil {
			t.Errorf("Call %d: unexpected error: %v", i, err)
		}
		if got != expected {
			t.Errorf("Call %d: got = %v, want = %v", i, got, expected)
		}
	}
}
