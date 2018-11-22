package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"fmt"
)

// Adapted version of https://github.com/looplab/tarjan/blob/master/tarjan.go

type tarjanData struct {
	tarjan
	nodes []tarjanNode
	stack []Step
	index map[string]int
}

type tarjan struct {
	graph  map[string]Step
	output [][]Step
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

	deps, err := td.graph[v].Dependencies()
	if err != nil {
		return nil, err
	}
	for w, _ := range deps {
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

func NewTarjan(steps map[string]Step) (*tarjan, error) {
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
	return &t.tarjan, nil
}

func (t *tarjan) Parse() (*[][]Step, error) {
	result := make([][]Step, 0)
	// walk reverse order, if all requirements are found the next step is a new component
	resultIndex := 0
	requirements := make(map[string]bool, 0)
	for i := len(t.output) - 1; i >= 0; i-- {
		steps := t.output[i]
		if len(steps) > 1 {
			return nil, fmt.Errorf("cyclic component found in pipeline: '%#v'", steps)
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
	return &result, nil
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
	t, err := NewTarjan(storage)
	if err != nil {
		return err
	}

	// write result
	result, err := t.Parse()
	*l = *result
	return err
}

type ServiceList [][]Step

func (l *ServiceList) UnmarshalJSON(data []byte) error {
	serviceStorage := make(map[string]Service, 0)
	stepStorage := make(map[string]Step, 0)
	err := json.Unmarshal(data, &serviceStorage)
	if err != nil {
		return err
	}
	for name, step := range serviceStorage {
		step.Name = name
		stepStorage[name] = Step{
			Service: step,
			Detach:  true,
		}
	}

	// Determine components and topological order
	t, err := NewTarjan(stepStorage)
	if err != nil {
		return err
	}

	// write result
	result, err := t.Parse()
	*l = *result
	return err
}
