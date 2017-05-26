package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"strings"
)

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
func (c *apiWikiCrawler) Fetch(ctx context.Context, link Link) (*Page, error) {
	//fmt.Printf("GET http://%s\n", link)
	//switch string(link) {
	//case "Mike Tyson":
	//	return &Page{
	//		Name: "Mike Tyson",
	//		Links: map[Link]uint64{"level2-1": 1, "level2-2":1},
	//	}, nil
	//case "level2-1":
	//	return &Page{
	//		Name: "level2-1",
	//		Links: map[Link]uint64{"level3-1": 1, "level3-2":1},
	//	}, nil
	//case "level2-2":
	//	return &Page{
	//		Name: "level2-2",
	//		Links: map[Link]uint64{"level3-3": 1, "level3-4":1},
	//	}, nil
	//case "level3-4":
	//	return &Page{
	//		Name: "level3-4",
	//		Links: map[Link]uint64{"level4-1":1},
	//	}, nil
	//case "level4-1":
	//	return &Page{
	//		Name: "level4-1",
	//		Links: map[Link]uint64{"Ukraine":1},
	//	}, nil
	//
	//default:
	//}
	//
	//return nil, fmt.Errorf("%s not found", link)

	page := &Page{
		Name:  link,
		Links: make(map[Link]uint64),
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
				link := Link(l.Title)

				if _, ok := page.Links[link]; !ok {
					page.Links[link] = 1
				} else {
					page.Links[link]++
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
