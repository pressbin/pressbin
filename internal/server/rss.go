package server

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	Items       []rssItem  `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func (s *Server) handleRSS(w http.ResponseWriter, r *http.Request) {
	pd := s.pageData()
	posts, _, err := s.store.ListPosts(1, 50)
	if err != nil {
		writeError(w, 500, "failed to list posts")
		return
	}
	base := pd.SiteURL
	if base == "" {
		base = "http://" + r.Host
	}
	items := make([]rssItem, 0, len(posts))
	for _, p := range posts {
		if p.Status != "published" {
			continue
		}
		link := fmt.Sprintf("%s/p/%s", trimSlash(base), p.Slug)
		items = append(items, rssItem{
			Title:       p.Title,
			Link:        link,
			Description: p.Summary,
			PubDate:     p.PublishedAt.UTC().Format(time.RFC1123Z),
			GUID:        link,
		})
	}
	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       pd.SiteTitle,
			Link:        trimSlash(base) + "/",
			Description: pd.SiteDescription,
			Items:       items,
		},
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(feed)
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
