package gantry // import "github.com/ad-freiburg/gantry"
// Adapted version of https://github.com/looplab/tarjan/blob/master/tarjan.go
import (
	"fmt"
)

type tarjanData struct {
	nodes  []tarjanNode
	stack  []Step
	index  map[string]int
	graph  map[string]Step
	output Pipelines
}

type tarjanNode struct {
	lowlink int
	stacked bool
}

func (td *tarjanData) strongConnect(v string) (*tarjanNode, error) {
	index := len(td.nodes)
	td.index[v] = index
	td.stack = append(td.stack, td.graph[v])
	td.nodes = append(td.nodes, tarjanNode{lowlink: index, stacked: true})
	node := &td.nodes[index]

	for w, _ := range *td.graph[v].Dependencies() {
		if _, ok := td.graph[w]; !ok {
			return nil, fmt.Errorf("Unknown dependency '%s' for step '%s'", w, v)
		}
		i, seen := td.index[w]
		if !seen {
			n, err := td.strongConnect(w)
			if err != nil {
				return nil, err
			}
			if n.lowlink < node.lowlink {
				node.lowlink = n.lowlink
			}
		} else if td.nodes[i].stacked {
			if i < node.lowlink {
				node.lowlink = i
			}
		}
	}

	if node.lowlink == index {
		var vertices []Step
		i := len(td.stack) - 1
		for {
			w := td.stack[i]
			stackIndex := td.index[w.Name]
			td.nodes[stackIndex].stacked = false
			vertices = append(vertices, w)
			if stackIndex == index {
				break
			}
			i--
		}
		td.stack = td.stack[:i]
		td.output = append(td.output, vertices)
	}
	return node, nil
}

func NewTarjan(steps map[string]Step) (*Pipelines, error) {
	// Determine components and topological order
	t := &tarjanData{
		nodes: make([]tarjanNode, 0, len(steps)),
		index: make(map[string]int, len(steps)),
	}
	t.graph = steps
	for v := range t.graph {
		if _, ok := t.index[v]; !ok {
			_, err := t.strongConnect(v)
			if err != nil {
				return nil, err
			}
		}
	}
	return &t.output, nil
}
