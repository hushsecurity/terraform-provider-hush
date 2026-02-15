package client

import (
	"context"
	"fmt"
	"time"
)

const (
	statusPollTimeout  = 3 * time.Minute
	statusPollInterval = 10 * time.Second
)

// statusResource is implemented by any resource type that has status fields.
type statusResource interface {
	statusFields() (status, statusDetail string)
}

// waitForResourceStatus polls the given getter until a terminal status is reached.
func waitForResourceStatus[T statusResource](ctx context.Context, c *Client, id string,
	getFn func(context.Context, *Client, string) (*T, error)) error {
	return waitForStatus(ctx, func() (string, string, error) {
		r, err := getFn(ctx, c, id)
		if err != nil {
			return "", "", err
		}
		status, detail := (*r).statusFields()
		return status, detail, nil
	})
}

func waitForStatus(ctx context.Context, pollFn func() (
	status, statusDetail string, err error)) error {
	deadline := time.Now().Add(statusPollTimeout)

	for {
		status, statusDetail, err := pollFn()
		if err != nil {
			return fmt.Errorf("error polling status: %w", err)
		}

		switch status {
		case "ok", "disabled":
			return nil
		case "warning", "error":
			return fmt.Errorf("resource entered %s status: %s", status, statusDetail)
		}

		// Non-terminal status (e.g. "syncing") — keep polling
		if time.Now().After(deadline) {
			return fmt.Errorf(
				"timed out waiting for resource to reach a terminal status (current: %s)",
				status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(statusPollInterval):
		}
	}
}
