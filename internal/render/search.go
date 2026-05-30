package render

import (
	"fmt"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func SearchResults(posts []store.Post, query string) g.Node {
	n := len(posts)
	if n == 0 {
		return h.Div(
			h.Class("pb-search-panel"),
			searchHeader(query, 0),
			h.P(h.Class("pb-muted"), g.Text("Try a different word or check the spelling.")),
		)
	}
	return h.Div(
		h.Class("pb-search-panel"),
		searchHeader(query, n),
		postList(posts, true),
	)
}

func searchHeader(query string, count int) g.Node {
	var label string
	switch count {
	case 0:
		label = fmt.Sprintf("No results for %q", query)
	case 1:
		label = fmt.Sprintf("1 result for %q", query)
	default:
		label = fmt.Sprintf("%d results for %q", count, query)
	}
	return h.P(h.Class("pb-search-header"), g.Text(label))
}
