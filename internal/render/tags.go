package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TagsPage(tags []string, pd PageData) g.Node {
	return Layout("Tags", pd,
		h.Div(
			h.Class("pb-stack pb-stack-lg"),
			h.H1(h.Class("pb-h1"), g.Text("Tags")),
			h.P(h.Class("pb-lead"), g.Text("All tags used on published posts.")),
			tagIndex(tags),
		),
	)
}

func tagIndex(tags []string) g.Node {
	if len(tags) == 0 {
		return h.P(h.Class("pb-muted"), g.Text("No tags yet."))
	}
	items := make([]g.Node, 0, len(tags))
	for _, t := range tags {
		items = append(items, h.A(
			h.Class("pb-pill"),
			h.Href("/tag/"+t),
			g.Text(t),
		))
	}
	return h.Div(h.Class("pb-pills"), g.Group(items))
}
