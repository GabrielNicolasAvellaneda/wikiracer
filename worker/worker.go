package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// WikiCrawler defines a wiki crawler interface
type WikiCrawler interface {
	Fetch(context.Context, Link) (*Page, error)
}

// NewHTMLWikiCrawler returns a new wiki crawler that parses html.
func NewHTMLWikiCrawler(client *http.Client) WikiCrawler {
	return &htmlWikiCrawler{
		client: client,
		endpoint: url.URL{
			Scheme: "https",
			Host:   "en.wikipedia.org",
			Path:   "/wiki/",
		},
	}
}

// htmlWikiCrawler parses wiki html page
type htmlWikiCrawler struct {
	client *http.Client

	endpoint url.URL
}

func (c *htmlWikiCrawler) trim(u string) Link {
	trimmed := strings.TrimLeft(u, "/wiki/")
	if index := strings.Index(u, "#"); index > -1 {
		trimmed = u[index:]
	}
	return Link(trimmed)
}

// Fetch gets a link and returns a *Page which represents a page with found links.
func (c *htmlWikiCrawler) Fetch(ctx context.Context, link Link) (*Page, error) {
	page := &Page{
		Name:  link,
		Links: make(map[Link]uint64),
	}

	pageURL := c.endpoint.String() + string(link)
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to make a new request: %s", err)
	}

	logrus.Debugf("GET %s", req.URL.String())
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("unable to do a request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response: %d", resp.StatusCode)
	}

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return page, nil
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			for _, a := range t.Attr {
				if a.Key == "href" {
					if !strings.HasPrefix(a.Val, "/wiki/") {
						continue
					}
					if strings.Contains(a.Val, ":") {
						continue
					}

					l := c.trim(a.Val)
					if _, ok := page.Links[l]; !ok {
						page.Links[l] = 1
					} else {
						page.Links[l]++
					}
				}

			}
		}
	}
}
