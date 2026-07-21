package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// newPagedClient returns a client wired to a mock whose single list endpoint
// pages with the given size, so by-name lookups must follow next_page.
func newPagedClient(t *testing.T, endpoint string, pageSize int) (*client.Client, *testutil.MockServer) {
	t.Helper()
	ms := testutil.NewMockServer(&testutil.Fixtures{
		Endpoints: map[string]map[string]any{endpoint: {}},
	})
	t.Cleanup(ms.Close)
	ms.SetPageSize(pageSize)
	c, err := client.NewClient(context.Background(), "mock-id", "mock-secret", ms.URL())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, ms
}

// TestGetDeploymentsByName_FollowsPagination proves collectPages walks every page
// of a cursor-paginated response across the interesting count boundaries. With
// the mock page size set to 2, any count above 2 forces the by-name lookup to
// follow next_page; before HUSH-6698 only the first page was read.
func TestGetDeploymentsByName_FollowsPagination(t *testing.T) {
	c, ms := newPagedClient(t, "GET /v1/deployments", 2)

	// A decoy under a different name guards against the server-side name filter
	// leaking unrelated rows into a paged result.
	ms.SeedObject("deployments", "dep-decoy", map[string]any{"id": "dep-decoy", "name": "decoy"})

	ctx := context.Background()
	// Cover single page, an exactly-full page, an exact multiple, and remainders.
	for _, total := range []int{1, 2, 4, 5, 7} {
		name := fmt.Sprintf("dup-%d", total)
		for i := 0; i < total; i++ {
			id := fmt.Sprintf("dep-%s-%02d", name, i)
			ms.SeedObject("deployments", id, map[string]any{"id": id, "name": name})
		}

		got, err := client.GetDeploymentsByName(ctx, c, name)
		if err != nil {
			t.Fatalf("GetDeploymentsByName(%q): %v", name, err)
		}
		if len(got) != total {
			t.Errorf("name %q: expected %d deployments across pages, got %d", name, total, len(got))
		}
		for _, d := range got {
			if d.Name != name {
				t.Errorf("name %q: filter leaked deployment %q (%s)", name, d.Name, d.ID)
			}
		}
	}
}

// countFn adapts a typed by-name lookup to (count, error) so every lookup can be
// exercised by one table.
type countFn func(context.Context, *client.Client, string) (int, error)

func adapt[T any](fn func(context.Context, *client.Client, string) ([]T, error)) countFn {
	return func(ctx context.Context, c *client.Client, v string) (int, error) {
		xs, err := fn(ctx, c, v)
		return len(xs), err
	}
}

// TestByNameLookups_FollowPagination exercises every paginated by-name/by-trigger
// lookup end-to-end: each must collect all matches across pages (page size 2) and
// keep excluding a decoy that shares the filter value. This catches per-lookup
// wiring mistakes -- wrong endpoint, wrong type= filter, or a wrong next_page tag
// on the list response -- that the shared collectPages unit test cannot.
func TestByNameLookups_FollowPagination(t *testing.T) {
	cases := []tableCase{
		{"gitlab integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "gitlab"}, integDecoy, adapt(client.GetGitlabIntegrationsByName)},
		{"confluence integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "confluence"}, integDecoy, adapt(client.GetConfluenceIntegrationsByName)},
		{"jira integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "jira"}, integDecoy, adapt(client.GetJiraIntegrationsByName)},
		{"bitbucket integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "bitbucket"}, integDecoy, adapt(client.GetBitbucketIntegrationsByName)},
		{"infisical integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "infisical"}, integDecoy, adapt(client.GetInfisicalIntegrationsByName)},
		{"sonatype integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "sonatype"}, integDecoy, adapt(client.GetSonatypeIntegrationsByName)},
		{"artifactory integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "artifactory"}, integDecoy, adapt(client.GetArtifactoryIntegrationsByName)},
		{"aws integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "aws"}, integDecoy, adapt(client.GetAWSIntegrationsByName)},
		{"gcp integration", "GET /v1/integrations", "integrations", "name",
			map[string]any{"type": "gcp"}, integDecoy, adapt(client.GetGCPIntegrationsByName)},
		{"deployment", "GET /v1/deployments", "deployments", "name",
			nil, nameDecoy, adapt(client.GetDeploymentsByName)},
		{"notification channel", "GET /v1/notification_channels", "notification_channels", "name",
			nil, nameDecoy, adapt(client.GetNotificationChannelsByName)},
		{"notification configuration by name", "GET /v1/notification_configurations", "notification_configurations", "name",
			nil, nameDecoy, adapt(client.GetNotificationConfigurationsByName)},
		{"notification configuration by trigger", "GET /v1/notification_configurations", "notification_configurations", "trigger",
			nil, map[string]any{"id": "decoy", "trigger": "other"}, adapt(client.GetNotificationConfigurationsByTrigger)},
		{"secret store", "GET /v1/secret_stores", "secret_stores", "name",
			nil, nameDecoy, adapt(client.GetSecretStoresByName)},
	}

	const total = 5 // > page size, so results span multiple pages
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			c, ms := newPagedClient(t, tc.endpoint, 2)
			for i := 0; i < total; i++ {
				id := fmt.Sprintf("row-%02d", i)
				obj := map[string]any{"id": id, tc.filterKey: "dup"}
				for k, v := range tc.itemFields {
					obj[k] = v
				}
				ms.SeedObject(tc.store, id, obj)
			}
			ms.SeedObject(tc.store, fmt.Sprintf("%v", tc.decoy["id"]), tc.decoy)

			n, err := tc.lookup(context.Background(), c, "dup")
			if err != nil {
				t.Fatalf("%s: lookup: %v", tc.label, err)
			}
			if n != total {
				t.Fatalf("%s: expected %d across pages (decoy excluded), got %d", tc.label, total, n)
			}
		})
	}
}

type tableCase struct {
	label      string
	endpoint   string
	store      string
	filterKey  string         // "name" or "trigger"
	itemFields map[string]any // extra fields on matching rows (e.g. type)
	decoy      map[string]any // a row that shares the filter value but must be excluded
	lookup     countFn
}

// nameDecoy differs by name; integDecoy shares the name but differs by type.
var (
	nameDecoy  = map[string]any{"id": "decoy", "name": "other"}
	integDecoy = map[string]any{"id": "decoy", "name": "dup", "type": "other"}
)
