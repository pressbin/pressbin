package render

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TagsPage(tags []string, pd PageData) g.Node {
	return Layout("Tags", pd,
		h.Div(
			h.Class("space-y-6"),
			h.H1(h.Class("text-3xl font-bold"), g.Text("Tags")),
			h.P(h.Class("text-gray-600"), g.Text("All tags used on published posts.")),
			tagIndex(tags),
		),
	)
}

func tagIndex(tags []string) g.Node {
	if len(tags) == 0 {
		return h.P(h.Class("text-gray-500"), g.Text("No tags yet."))
	}
	items := make([]g.Node, 0, len(tags))
	for _, t := range tags {
		items = append(items, h.A(
			h.Class("inline-block px-3 py-1 rounded-full bg-gray-100 text-gray-800 hover:bg-blue-100 hover:text-blue-800 text-sm"),
			h.Href("/tag/"+t),
			g.Text(t),
		))
	}
	return h.Div(h.Class("flex flex-wrap gap-2"), g.Group(items))
}
