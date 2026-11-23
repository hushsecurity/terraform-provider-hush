package cloudflare_r2_access_credential

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestKeepPriorState(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"refused PATCH", &client.APIError{StatusCode: 409}, true},
		{"token refresh failed", fmt.Errorf("token refresh failed: %w", errors.New("boom")), true},
		{"transport failure", fmt.Errorf("do request: %w", errors.New("reset")), true},
		{"failed status wait", &client.StatusWaitError{Err: errors.New("resource entered error status")}, false},
		{"failed status poll", &client.StatusWaitError{
			Err: fmt.Errorf("error polling status: %w", &client.APIError{StatusCode: 500})}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := keepPriorState(tc.err); got != tc.want {
				t.Errorf("keepPriorState(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
