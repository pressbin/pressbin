package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func PostPage(post store.Post, pd PageData) g.Node {
	return Layout(post.Title, pd,
		h.Article(
			h.H1(h.Class("text-3xl font-bold mb-2"), g.Text(post.Title)),
			h.Div(
				h.Class("text-gray-500 text-sm mb-8 flex flex-wrap items-center gap-2"),
				g.Text(post.PublishedAt.Format("January 2, 2006")),
				tagList(post.Tags),
			),
			h.Div(
				h.Class("prose prose-gray max-w-none"),
				g.Raw(post.ContentHTML),
			),
		),
	)
}

func tagList(tags []string) g.Node {
	if len(tags) == 0 {
		return g.Group(nil)
	}
	nodes := make([]g.Node, 0, len(tags)*2)
	for i, t := range tags {
		if i > 0 {
			nodes = append(nodes, g.Text(" · "))
		}
		nodes = append(nodes, h.A(
			h.Class("text-blue-600 hover:underline"),
			h.Href("/tag/"+t),
			g.Text(t),
		))
	}
	return h.Span(h.Class("text-gray-400"), g.Group(nodes))
}
