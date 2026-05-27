package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func PostPage(post store.Post, pd PageData) g.Node {
	return Layout(post.Title, pd,
		h.Article(
			h.Header(
				h.Class("pb-post-header"),
				h.H1(h.Class("pb-post-h1"), g.Text(post.Title)),
				h.Div(
					h.Class("pb-post-meta"),
					g.Text(post.PublishedAt.Format("January 2, 2006")),
					tagList(post.Tags),
				),
			),
			h.Div(
				h.Class("pb-prose"),
				g.Raw(post.ContentHTML),
			),
		),
	)
}

func tagList(tags []string) g.Node {
	if len(tags) == 0 {
		return g.Group(nil)
	}
	nodes := make([]g.Node, 0, len(tags))
	for _, t := range tags {
		nodes = append(nodes, h.A(
			h.Class("pb-tag"),
			h.Href("/tag/"+t),
			g.Text(t),
		))
	}
	return h.Span(h.Class("pb-tags"), g.Group(nodes))
}
