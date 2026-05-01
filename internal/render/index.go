package render

import (
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
	hx "maragu.dev/gomponents-htmx"

	"pressbin.in/pressbin/internal/store"
)

func IndexPage(posts []store.Post, pd PageData, page, totalPages int) g.Node {
	return Layout(pd.SiteTitle, pd,
		h.Div(
			h.Class("space-y-10"),
			h.Div(
				h.Class("space-y-4"),
				h.H1(h.Class("text-3xl font-bold"), g.Text(pd.SiteTitle)),
				h.P(h.Class("text-gray-600"), g.Text(pd.SiteDescription)),
				searchBox(),
				h.Div(h.ID("search-results")),
			),
			postList(posts, false),
			pagination("/", page, totalPages),
		),
	)
}

func searchBox() g.Node {
	return h.Form(
		h.Class("w-full"),
		hx.Get("/search"),
		hx.Target("#search-results"),
		hx.Trigger("keyup changed delay:300ms from:input[name='q']"),
		h.Input(
			h.Type("search"),
			h.Name("q"),
			h.Class("w-full border border-gray-300 rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"),
			h.Placeholder("Search posts..."),
		),
	)
}

func postList(posts []store.Post, compact bool) g.Node {
	if len(posts) == 0 {
		return h.P(h.Class("text-gray-500"), g.Text("No posts yet."))
	}
	items := make([]g.Node, 0, len(posts))
	for _, p := range posts {
		items = append(items, postCard(p, compact))
	}
	return h.Div(h.Class("space-y-8"), g.Group(items))
}

func postCard(p store.Post, compact bool) g.Node {
	href := "/p/" + p.Slug
	summary := p.Summary
	if summary == "" && compact {
		summary = trimWords(stripHTML(p.ContentHTML), 40)
	}
	cls := "block group"
	if compact {
		cls += " py-3 border-b border-gray-100 last:border-0"
	}
	return h.A(
		h.Class(cls),
		h.Href(href),
		h.H2(h.Class("text-xl font-semibold text-gray-900 group-hover:text-blue-600"), g.Text(p.Title)),
		h.Div(
			h.Class("text-sm text-gray-500 mt-1"),
			g.Text(p.PublishedAt.Format("Jan 2, 2006")),
			tagList(p.Tags),
		),
		g.If(summary != "", h.P(h.Class("text-gray-600 mt-2 line-clamp-2"), g.Text(summary))),
	)
}

func pagination(basePath string, page, totalPages int) g.Node {
	if totalPages <= 1 {
		return g.Group(nil)
	}
	var nodes []g.Node
	if page > 1 {
		prev := basePath
		if page-1 > 1 {
			prev = basePath + "?page=" + strconv.Itoa(page-1)
		}
		nodes = append(nodes, h.A(h.Class("text-blue-600 hover:underline"), h.Href(prev), g.Text("← Newer")))
	}
	nodes = append(nodes, h.Span(h.Class("text-gray-500 text-sm"), g.Text("Page "+strconv.Itoa(page)+" of "+strconv.Itoa(totalPages))))
	if page < totalPages {
		nodes = append(nodes, h.A(h.Class("text-blue-600 hover:underline"), h.Href(basePath+"?page="+strconv.Itoa(page+1)), g.Text("Older →")))
	}
	return h.Div(h.Class("flex justify-between items-center pt-8 border-t border-gray-200"), g.Group(nodes))
}
