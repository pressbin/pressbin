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
	CustomCSSURL    string
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
			h.Script(h.Src("https://unpkg.com/htmx.org@1.9.12")),
			h.Link(h.Rel("preconnect"), h.Href("https://fonts.googleapis.com")),
			h.Link(h.Rel("preconnect"), h.Href("https://fonts.gstatic.com"), h.CrossOrigin("")),
			h.Link(h.Rel("stylesheet"), h.Href("https://fonts.googleapis.com/css2?family=DM+Mono:ital,wght@0,400;0,500;1,400&family=Instrument+Serif:ital@0;1&family=DM+Sans:opsz,wght@9..40,400;9..40,500;9..40,600&display=swap")),
			h.Link(h.Rel("stylesheet"), h.Href("/theme/style.css")),
			g.If(strings.TrimSpace(pd.CustomCSSURL) != "",
				h.Link(h.Rel("stylesheet"), h.Href(pd.CustomCSSURL)),
			),
			h.Link(h.Rel("alternate"), h.Type("application/rss+xml"), h.Href("/feed.xml")),
		),
		h.Body(
			h.Class("pb-body"),
			navbar(pd),
			h.Main(h.Class("pb-main"), content),
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
		h.Class("pb-nav"),
		h.Div(
			h.Class("pb-nav-inner"),
			h.A(
				h.Class("pb-nav-logo"),
				h.Href(home),
				g.Text(pd.SiteTitle),
			),
			h.Div(
				h.Class("pb-nav-links"),
				h.A(
					h.Class("pb-nav-link"),
					h.Href(home),
					lucide.House(h.Class("pb-icon")),
					g.Text("Home"),
				),
				h.A(
					h.Class("pb-nav-link"),
					h.Href("/tags"),
					g.Text("Tags"),
				),
				h.A(
					h.Class("pb-nav-link"),
					h.Href("/feed.xml"),
					lucide.Rss(h.Class("pb-icon")),
					g.Text("RSS"),
				),
			),
		),
	)
}

func footer() g.Node {
	return h.Footer(
		h.Class("pb-footer"),
		g.Text("Powered by Pressbin"),
	)
}
