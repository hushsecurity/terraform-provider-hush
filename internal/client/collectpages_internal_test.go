package client

import (
	"errors"
	"reflect"
	"testing"
)

// TestCollectPages covers the shared paging loop used by every by-name/by-trigger
// lookup, independent of any endpoint: accumulation, cursor threading, both stop
// conditions, and error propagation.
func TestCollectPages(t *testing.T) {
	sp := func(s string) *string { return &s }

	t.Run("accumulates across pages and threads the cursor", func(t *testing.T) {
		pages := []struct {
			items []int
			next  *string
		}{
			{[]int{1, 2}, sp("c1")},
			{[]int{3, 4}, sp("c2")},
			{[]int{5}, nil},
		}
		var seenCursors []string
		i := 0
		got, err := collectPages(func(cursor string) ([]int, *string, error) {
			seenCursors = append(seenCursors, cursor)
			p := pages[i]
			i++
			return p.items, p.next, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
			t.Fatalf("items = %v, want %v", got, want)
		}
		if want := []string{"", "c1", "c2"}; !reflect.DeepEqual(seenCursors, want) {
			t.Fatalf("cursors passed = %v, want %v", seenCursors, want)
		}
	})

	t.Run("stops on nil next", func(t *testing.T) {
		calls := 0
		got, err := collectPages(func(string) ([]int, *string, error) {
			calls++
			return []int{7}, nil, nil
		})
		if err != nil || calls != 1 || !reflect.DeepEqual(got, []int{7}) {
			t.Fatalf("got %v, err %v, calls %d", got, err, calls)
		}
	})

	t.Run("stops on empty-string next", func(t *testing.T) {
		calls := 0
		got, err := collectPages(func(string) ([]int, *string, error) {
			calls++
			return []int{8}, sp(""), nil
		})
		if err != nil || calls != 1 || !reflect.DeepEqual(got, []int{8}) {
			t.Fatalf("got %v, err %v, calls %d", got, err, calls)
		}
	})

	t.Run("propagates fetch error", func(t *testing.T) {
		sentinel := errors.New("boom")
		_, err := collectPages(func(string) ([]int, *string, error) {
			return nil, nil, sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("err = %v, want %v", err, sentinel)
		}
	})
}

// TestWithCursor covers the query-append edge that the by-name lookups rely on:
// a filtered path (name=/type=) must gain the cursor with & , a bare path with ?,
// and the cursor must be escaped.
func TestWithCursor(t *testing.T) {
	cases := []struct {
		name, path, cursor, want string
	}{
		{"empty cursor unchanged", "/v1/deployments?name=x", "", "/v1/deployments?name=x"},
		{"appends with & when query present", "/v1/integrations?name=x&type=gitlab", "c1", "/v1/integrations?name=x&type=gitlab&cursor=c1"},
		{"appends with ? when no query", "/v1/deployments", "c1", "/v1/deployments?cursor=c1"},
		{"escapes the cursor", "/v1/x?y=1", "a b/c", "/v1/x?y=1&cursor=a+b%2Fc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := withCursor(tc.path, tc.cursor); got != tc.want {
				t.Fatalf("withCursor(%q, %q) = %q, want %q", tc.path, tc.cursor, got, tc.want)
			}
		})
	}
}
