package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func SearchResults(posts []store.Post, query string) g.Node {
	if query == "" {
		return g.Group(nil)
	}
	if len(posts) == 0 {
		return h.Div(
			h.Class("pb-search-empty"),
			h.P(h.Class("pb-muted"), g.Text("No results for \""+query+"\".")),
		)
	}
	return h.Div(
		h.Class("pb-search-results"),
		h.P(h.Class("pb-muted pb-text-sm"), g.Text("Results for \""+query+"\"")),
		postList(posts, true),
	)
}
