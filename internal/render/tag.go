package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func TagPage(tag string, posts []store.Post, pd PageData, page, totalPages int) g.Node {
	title := "Tag: " + tag
	return Layout(title, pd,
		h.Div(
			h.Class("space-y-8"),
			h.H1(h.Class("text-3xl font-bold"), g.Text(title)),
			postList(posts, false),
			pagination("/tag/"+tag, page, totalPages),
		),
	)
}
