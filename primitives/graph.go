package primitives

import graph "gopkg.in/gyuho/goraph.v2"

// Graph is a graph interface based on gopkg.in/gyuho/goraph.v2
type Graph interface {
	graph.Graph

	GAddNode(n string) bool
	GAddEdge(id1, id2 string, w float64) error
	GFindPath(n1, n2 string) ([]string, error)
	GBFS(id string) []string
}

type goraphImpl struct {
	graph.Graph
}

func (g *goraphImpl) GBFS(id string) []string {
	ids := graph.BFS(g, graph.StringID(id))
	result := make([]string, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

// GAddNode adds a new node to graph.
func (g *goraphImpl) GAddNode(n string) bool {
	node := graph.NewNode(n)
	return g.AddNode(node)
}

// GAddEdge adds a new edge between id1 and id2 with weight w.
func (g *goraphImpl) GAddEdge(id1, id2 string, w float64) error {
	return g.AddEdge(graph.StringID(id1), graph.StringID(id2), w)
}

// GFindPath finds a path between n1 and n2.
func (g *goraphImpl) GFindPath(n1, n2 string) ([]string, error) {
	path, _, err := graph.Dijkstra(g, graph.StringID(n1), graph.StringID(n2))
	var result []string
	for _, node := range path {
		result = append(result, node.String())
	}

	return result, err
}

// NewGraph returns a new graph.
func NewGraph() Graph {
	return &goraphImpl{
		graph.NewGraph(),
	}
}
