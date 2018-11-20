package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"fmt"
)

// Adapted version of https://github.com/looplab/tarjan/blob/master/tarjan.go

type tarjan struct {
	graph  map[string]Step
	nodes  []tarjanNode
	stack  []Step
	index  map[string]int
	output [][]Step
}

type tarjanNode struct {
	lowlink int
	stacked bool
}

func (tarjan *tarjan) strongConnect(v string) (*tarjanNode, error) {
	index := len(tarjan.nodes)
	tarjan.index[v] = index
	tarjan.stack = append(tarjan.stack, tarjan.graph[v])
	tarjan.nodes = append(tarjan.nodes, tarjanNode{lowlink: index, stacked: true})
	node := &tarjan.nodes[index]

	for w, _ := range tarjan.graph[v].After {
		if _, ok := tarjan.graph[w]; !ok {
			return nil, fmt.Errorf("Unknown dependency '%s' for step '%s'", w, v)
		}
		i, seen := tarjan.index[w]
		if !seen {
			n, err := tarjan.strongConnect(w)
			if err != nil {
				return nil, err
			}
			if n.lowlink < node.lowlink {
				node.lowlink = n.lowlink
			}
		} else if tarjan.nodes[i].stacked {
			if i < node.lowlink {
				node.lowlink = i
			}
		}
	}

	if node.lowlink == index {
		var vertices []Step
		i := len(tarjan.stack) - 1
		for {
			w := tarjan.stack[i]
			stackIndex := tarjan.index[w.Name]
			tarjan.nodes[stackIndex].stacked = false
			vertices = append(vertices, w)
			if stackIndex == index {
				break
			}
			i--
		}
		tarjan.stack = tarjan.stack[:i]
		tarjan.output = append(tarjan.output, vertices)
	}
	return node, nil
}

type StepList [][]Step

func (l StepList) All() []Step {
	result := make([]Step, 0)
	for _, steps := range l {
		result = append(result, steps...)
	}
	return result
}

func (l *StepList) UnmarshalJSON(data []byte) error {
	result := make([][]Step, 0)

	storage := make(map[string]Step, 0)
	err := json.Unmarshal(data, &storage)
	if err != nil {
		return err
	}
	for name, step := range storage {
		step.Name = name
		storage[name] = step
	}

	// Determine components and topological order
	t := &tarjan{
		graph: storage,
		nodes: make([]tarjanNode, 0, len(storage)),
		index: make(map[string]int, len(storage)),
	}
	for v := range t.graph {
		if _, ok := t.index[v]; !ok {
			_, err := t.strongConnect(v)
			if err != nil {
				return err
			}
		}
	}

	// walk reverse order, if all requirements are found the next step is a new component
	resultIndex := 0
	requirements := make(map[string]bool, 0)
	for i := len(t.output) - 1; i >= 0; i-- {
		steps := t.output[i]
		if len(steps) > 1 {
			return fmt.Errorf("cyclic component found in pipeline: '%#v'", steps)
		}
		var step = steps[0]
		for r, _ := range step.After {
			requirements[r] = true
		}
		delete(requirements, step.Name)
		if len(result)-1 < resultIndex {
			result = append(result, make([]Step, 0))
		}
		result[resultIndex] = append([]Step{step}, result[resultIndex]...)
		if len(requirements) == 0 {
			resultIndex++
		}
	}

	// write result
	*l = result
	return nil
}
