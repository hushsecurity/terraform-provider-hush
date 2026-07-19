package client

import (
	"net/url"
	"strings"
)

// collectPages follows a cursor-paginated list endpoint and returns the
// concatenated items across every page. fetchPage is called once per page with
// the cursor for the page to fetch (empty on the first request) and returns that
// page's items and the cursor for the next page (nil/empty when exhausted).
//
// The backend uses hush.pagination.CursorPage everywhere: the request carries a
// `cursor` query parameter and each response reports the next page in `next_page`.
func collectPages[T any](fetchPage func(cursor string) (items []T, next *string, err error)) ([]T, error) {
	var all []T
	cursor := ""
	for {
		items, next, err := fetchPage(cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if next == nil || *next == "" {
			break
		}
		cursor = *next
	}
	return all, nil
}

// withCursor appends the pagination cursor to a list path that may already carry
// filter query parameters.
func withCursor(path, cursor string) string {
	if cursor == "" {
		return path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "cursor=" + url.QueryEscape(cursor)
}
