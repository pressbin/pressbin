package render

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
	"github.com/eduardolat/gomponents-lucide"
)

type PageData struct {
	SiteTitle       string
	SiteURL         string
	SiteDescription string
}

func Layout(title string, pd PageData, content g.Node) g.Node {
	fullTitle := title
	if pd.SiteTitle != "" && title != "" && title != pd.SiteTitle {
		fullTitle = title + " — " + pd.SiteTitle
	} else if title == "" {
		fullTitle = pd.SiteTitle
	}
	return h.HTML(
		h.Lang("en"),
		h.Head(
			h.Meta(h.Charset("utf-8")),
			h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
			h.Meta(h.Name("description"), h.Content(pd.SiteDescription)),
			h.TitleEl(g.Text(fullTitle)),
			h.Script(h.Src("https://cdn.tailwindcss.com")),
			h.Script(h.Src("https://unpkg.com/htmx.org@1.9.12")),
			h.Link(h.Rel("stylesheet"), h.Href("/assets/style.css")),
			h.Link(h.Rel("alternate"), h.Type("application/rss+xml"), h.Href("/feed.xml")),
		),
		h.Body(
			h.Class("bg-gray-50 text-gray-900 min-h-screen"),
			navbar(pd),
			h.Main(h.Class("max-w-2xl mx-auto px-4 py-12"), content),
			footer(),
		),
	)
}

func navbar(pd PageData) g.Node {
	home := "/"
	if base := strings.TrimSuffix(strings.TrimSpace(pd.SiteURL), "/"); base != "" {
		home = base + "/"
	}
	return h.Nav(
		h.Class("border-b border-gray-200 bg-white"),
		h.Div(
			h.Class("max-w-2xl mx-auto px-4 py-4 flex items-center justify-between gap-4"),
			h.A(
				h.Class("text-lg font-semibold text-gray-900 hover:text-blue-600"),
				h.Href(home),
				g.Text(pd.SiteTitle),
			),
			h.Div(
				h.Class("flex items-center gap-4 text-sm"),
				h.A(
					h.Class("inline-flex items-center gap-1 text-gray-600 hover:text-blue-600"),
					h.Href(home),
					lucide.House(h.Class("w-4 h-4")),
					g.Text("Home"),
				),
				h.A(
					h.Class("text-gray-600 hover:text-blue-600"),
					h.Href("/tags"),
					g.Text("Tags"),
				),
				h.A(
					h.Class("inline-flex items-center gap-1 text-gray-600 hover:text-blue-600"),
					h.Href("/feed.xml"),
					lucide.Rss(h.Class("w-4 h-4")),
					g.Text("RSS"),
				),
			),
		),
	)
}

func footer() g.Node {
	return h.Footer(
		h.Class("max-w-2xl mx-auto px-4 py-8 text-center text-sm text-gray-500 border-t border-gray-200 mt-12"),
		g.Text("Powered by Pressbin"),
	)
}
