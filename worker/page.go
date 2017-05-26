package worker

// Page represents a crawled wikipedia page with Name and embedded Links.
type Page struct {
	Name  string
	Depth int
	Prev *Page
	Links map[string]bool
}
