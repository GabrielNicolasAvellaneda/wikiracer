package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// WikiCrawler defines a wiki crawler interface
type WikiCrawler interface {
	Fetch(context.Context, string) (*Page, error)
}

// NewAPIWikiCrawler is a apiWikiCrawler constructor.
func NewAPIWikiCrawler(client *http.Client) WikiCrawler {
	return &apiWikiCrawler{
		client: client,
		endpoint: url.URL{
			Scheme: "https",
			Host:   "en.wikipedia.org",
			Path:   "/w/api.php",
		},
	}
}

// apiWikiCrawler is using wikipedia api (en.wikipedia.org/w/api.php) to get the pages.
type apiWikiCrawler struct {
	client *http.Client

	endpoint url.URL
}

// Fetch takes a wiki Link and returns wiki Page.
func (c *apiWikiCrawler) Fetch(ctx context.Context, link string) (*Page, error) {
	page := &Page{
		Name:  link,
		Links: make(map[string]bool),
	}
	// response describes the response from the server.
	type response struct {
		Continue struct {
			Plcontinue string `json:"plcontinue"`
			Continue   string `json:"continue"`
		} `json:"continue"`

		Query struct {
			Pages map[string]struct {
				Links []struct {
					Title string `json:"title"`
				} `json:"links"`
			} `json:"pages"`
		} `json:"query"`
	}

	var cont string
	for {
		v := url.Values{}
		v.Add("action", "query")
		v.Add("format", "json")
		v.Add("prop", "links")
		v.Add("pllimit", "500")
		v.Add("titles", string(link))
		if cont != "" {
			v.Add("plcontinue", cont)
		}

		wikiURL := c.endpoint
		wikiURL.RawQuery = v.Encode()
		logrus.Debugf("GET %s", wikiURL.String())

		req, err := http.NewRequest("GET", wikiURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("unable to make a new request: %s", err)
		}

		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("bad response: %d", resp.StatusCode)
		}

		r := &response{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("unable to read response: %s", err)
		}
		resp.Body.Close()

		if err := json.Unmarshal(body, r); err != nil {
			return nil, fmt.Errorf("unable to unmarshal response: %s", err)
		}

		for _, p := range r.Query.Pages {
			for _, l := range p.Links {
				if strings.Contains(l.Title, ":") {
					continue
				}

				if _, ok := page.Links[l.Title]; !ok {
					page.Links[l.Title] = true
				}
			}
		}

		if r.Continue.Plcontinue == "" {
			break
		}
		cont = r.Continue.Plcontinue
	}

	return page, nil
}
