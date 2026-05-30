package render

import (
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	h "maragu.dev/gomponents/html"

	"pressbin.dev/pressbin/internal/store"
)

func IndexPage(posts []store.Post, pd PageData, page, totalPages int) g.Node {
	return Layout(pd.SiteTitle, pd,
		h.Div(
			h.Class("pb-stack pb-stack-lg"),
			h.Div(
				h.Class("pb-stack"),
				// Avoid repeating the site title (navbar already shows it).
				g.If(pd.SiteDescription != "", h.P(h.Class("pb-lead"), g.Text(pd.SiteDescription))),
				searchBox(),
				postFeed(posts, page, totalPages),
			),
		),
	)
}

func searchBox() g.Node {
	return h.Form(
		h.Class("pb-search"),
		hx.Get("/search"),
		hx.Target("#post-feed"),
		hx.Trigger("keyup changed delay:300ms from:input[name='q']"),
		h.Input(
			h.Type("search"),
			h.Name("q"),
			h.Class("pb-input"),
			h.Placeholder("Search posts..."),
		),
	)
}

func PostFeedFragment(posts []store.Post, page, totalPages int) g.Node {
	return postFeedInner(posts, page, totalPages)
}

func postFeed(posts []store.Post, page, totalPages int) g.Node {
	return h.Div(
		h.ID("post-feed"),
		postFeedInner(posts, page, totalPages),
	)
}

func postFeedInner(posts []store.Post, page, totalPages int) g.Node {
	return g.Group([]g.Node{
		postList(posts, false),
		pagination("/", page, totalPages),
	})
}

func postList(posts []store.Post, compact bool) g.Node {
	if len(posts) == 0 {
		return h.P(h.Class("pb-muted"), g.Text("No posts yet."))
	}
	items := make([]g.Node, 0, len(posts))
	for _, p := range posts {
		items = append(items, postCard(p, compact))
	}
	return h.Div(h.Class("pb-posts"), g.Group(items))
}

func postCard(p store.Post, compact bool) g.Node {
	href := "/p/" + p.Slug
	summary := p.Summary
	if summary == "" && compact {
		summary = trimWords(stripHTML(p.ContentHTML), 40)
	}
	cls := "pb-post-card"
	if compact {
		cls += " pb-post-card-compact"
	}
	return h.Div(
		h.Class(cls),
		h.H2(
			h.Class("pb-post-title"),
			h.A(h.Href(href), g.Text(p.Title)),
		),
		h.Div(
			h.Class("pb-post-meta"),
			g.Text(p.PublishedAt.Format("Jan 2, 2006")),
			tagList(p.Tags),
		),
		g.If(summary != "", h.P(h.Class("pb-post-summary"), g.Text(summary))),
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
		nodes = append(nodes, h.A(h.Class("pb-link"), h.Href(prev), g.Text("← Newer")))
	}
	nodes = append(nodes, h.Span(h.Class("pb-muted pb-text-sm"), g.Text("Page "+strconv.Itoa(page)+" of "+strconv.Itoa(totalPages))))
	if page < totalPages {
		nodes = append(nodes, h.A(h.Class("pb-link"), h.Href(basePath+"?page="+strconv.Itoa(page+1)), g.Text("Older →")))
	}
	return h.Div(h.Class("pb-pagination"), g.Group(nodes))
}
