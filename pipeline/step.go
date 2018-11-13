package pipeline // import "github.com/ad-freiburg/gantry/pipeline"

type Step struct {
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	After   StringSet `json:"after"`
	machine *Machine
}
