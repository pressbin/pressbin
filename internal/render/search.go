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
			h.Class("mt-4 p-4 bg-white border border-gray-200 rounded-lg"),
			h.P(h.Class("text-gray-600"), g.Text("No results for \""+query+"\".")),
		)
	}
	return h.Div(
		h.Class("mt-4 space-y-2"),
		h.P(h.Class("text-sm text-gray-500"), g.Text("Results for \""+query+"\"")),
		postList(posts, true),
	)
}
