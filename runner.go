package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"os/user"
	"strings"
	"sync"
)

const docker string = "docker"
const wharfer string = "wharfer"

func getContainerExecutable() string {
	if ForceWharfer {
		return wharfer
	}
	if isWharferInstalled() {
		if isUserRoot() || isUserInDockerGroup() {
			return docker
		}
		return wharfer
	}
	return docker
}

func isUserRoot() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Uid == "0"
}

func isUserInDockerGroup() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	gids, err := u.GroupIds()
	if err != nil {
		return false
	}
	for _, gid := range gids {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			return false
		}
		if group.Name == docker {
			return true
		}
	}
	return false
}

func isWharferInstalled() bool {
	cmd := exec.Command(wharfer, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// Runner represents generic container runners.
type Runner interface {
	Copy() Runner
	ImageBuilder(Step, bool) func() error
	ImagePuller(Step) func() error
	ImageExistenceChecker(Step) func() error
	ContainerKiller(Step) func() (int, error)
	ContainerRemover(Step) func() error
	ContainerRunner(Step, Network) func() error
	NetworkCreator(Network) func() error
	NetworkRemover(Network) func() error
}

// NoopRunner is a runner that does nothing.
type NoopRunner struct {
	silent bool
	calls  map[string]int
	called map[string]int
	mutex  sync.RWMutex
}

// NewLocalRunner returns a NoopRunner.
func NewNoopRunner(silent bool) *NoopRunner {
	return &NoopRunner{
		silent: silent,
		calls:  make(map[string]int),
		called: make(map[string]int),
	}
}

// Copy returns the same instance.
func (r *NoopRunner) Copy() Runner {
	return r
}

// NumCalls returns how many functions with the given key were created.
func (r *NoopRunner) NumCalls(key string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.calls[key]
}

// NumCalled returns how many functions with the given key were executed.
func (r *NoopRunner) NumCalled(key string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.called[key]
}

func (r *NoopRunner) incrementCalls(key string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.calls[key]++
}

func (r *NoopRunner) incrementCalled(key string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.called[key]++
}

// ImageBuilder returns a function to build the image for the given step.
func (r *NoopRunner) ImageBuilder(step Step, force bool) func() error {
	key := fmt.Sprintf("ImageBuilder(%s,%t)", step.Name, force)
	r.incrementCalls(key)
	return func() error {
		if !r.silent {
			pipelineLogger.Printf("- Building: %s!", step.ColoredContainerName())
		}
		r.incrementCalled(key)
		return nil
	}
}

// ImagePuller returns a function to pull the image for the given step.
func (r *NoopRunner) ImagePuller(step Step) func() error {
	key := fmt.Sprintf("ImagePuller(%s)", step.Name)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		return nil
	}
}

// ImageExistenceChecker returns a function which checks if the image for the given step exists.
func (r *NoopRunner) ImageExistenceChecker(step Step) func() error {
	key := fmt.Sprintf("ImageExistenceChecker(%s)", step.Name)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		return nil
	}
}

// ContainerKiller returns a function to kill the container for the given step.
func (r *NoopRunner) ContainerKiller(step Step) func() (int, error) {
	key := fmt.Sprintf("ContainerKiller(%s)", step.Name)
	r.incrementCalls(key)
	return func() (int, error) {
		r.incrementCalled(key)
		return 0, nil
	}
}

// ContainerRemover returns a function to remove the container for the given step.
func (r *NoopRunner) ContainerRemover(step Step) func() error {
	key := fmt.Sprintf("ContainerRemover(%s)", step.Name)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		return nil
	}
}

// ContainerRunner returns a function to run the given step.
func (r *NoopRunner) ContainerRunner(step Step, network Network) func() error {
	key := fmt.Sprintf("ContainerRunner(%s,%s)", step.Name, network)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		if !r.silent {
			pipelineLogger.Printf("- Skipping: %s!", step.ColoredContainerName())
		}
		return nil
	}
}

// NetworkCreator returns a function to create the given network.
func (r *NoopRunner) NetworkCreator(network Network) func() error {
	key := fmt.Sprintf("NetworkCreator(%s)", network)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		return nil
	}
}

// NetworkRemover returns a function to create the given network.
func (r *NoopRunner) NetworkRemover(network Network) func() error {
	key := fmt.Sprintf("NetworkRemover(%s)", network)
	r.incrementCalls(key)
	return func() error {
		r.incrementCalled(key)
		return nil
	}
}

// LocalRunner creates functions running on localhost.
type LocalRunner struct {
	prefix string
	stdout io.Writer
	stderr io.Writer
}

// NewLocalRunner returns a LocalRunner using provided defaults.
func NewLocalRunner(prefix string, stdout io.Writer, stderr io.Writer) *LocalRunner {
	return &LocalRunner{
		prefix: prefix,
		stdout: stdout,
		stderr: stderr,
	}
}

// Copy returns a new Instance with copied values.
func (r *LocalRunner) Copy() Runner {
	return &LocalRunner{
		prefix: r.prefix,
		stdout: r.stdout,
		stderr: r.stderr,
	}
}

// Exec executes given arguments with the containerExecutable.
func (r *LocalRunner) Exec(args []string) error {
	ce := getContainerExecutable()
	if ShowContainerCommands {
		log.Printf("Exec:   %s %s", ce, strings.Join(args, " "))
	}
	cmd := exec.Command(ce, args...)
	stdout := NewPrefixedLogger(r.prefix, log.New(r.stdout, "", log.LstdFlags))
	stderr := NewPrefixedLogger(r.prefix, log.New(r.stderr, "", log.LstdFlags))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// Output executes given arguments with the containerExecutable and returns the output.
func (r *LocalRunner) Output(args []string) ([]byte, error) {
	ce := getContainerExecutable()
	if ShowContainerCommands {
		log.Printf("Output: %s %s", ce, strings.Join(args, " "))
	}
	cmd := exec.Command(ce, args...)
	return cmd.Output()
}

// ImageBuilder returns a function to build the image for the given step.
func (r *LocalRunner) ImageBuilder(step Step, pull bool) func() error {
	return func() error {
		if Verbose {
			log.Printf("Build image for '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.BuildCommand(pull))
	}
}

// ImagePuller retunrs a function to pull the image for the given step.
func (r *LocalRunner) ImagePuller(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Pull image for '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.PullCommand())
	}
}

// ImageExistenceChecker returns a function which checks if the image for the given step exists.
func (r *LocalRunner) ImageExistenceChecker(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Check image ('%s') existence for '%s'", step.ImageName(), step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		// Search for image
		out, err := r.Output([]string{"images", "--format", "{{.ID}};{{.Repository}}", step.ImageName()})
		if err != nil {
			return err
		}
		// Count found images
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		count := 0
		for scanner.Scan() {
			count++
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		// If no image was found, raise error
		if count == 0 {
			return fmt.Errorf("image not found '%s'", step.ImageName())
		}
		return nil
	}
}

// ContainerKiller returns a function to kill the container for the given step.
func (r *LocalRunner) ContainerKiller(step Step) func() (int, error) {
	return func() (int, error) {
		var counter int
		if Verbose {
			log.Printf("Kill container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		// Get id(s) of container with name of step to kill
		out, err := r.Output([]string{"ps", "-q", "--filter", fmt.Sprintf("name=%s$", step.ContainerName())})
		if err != nil {
			return counter, err
		}
		// Kill all found containers
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			counter++
			if err := r.Exec([]string{"kill", scanner.Text()}); err != nil {
				return counter, err
			}
		}
		return counter, scanner.Err()
	}
}

// ContainerRemover returns a function to remove the container for the given step.
func (r *LocalRunner) ContainerRemover(step Step) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		// Get id(s) of container with name of step to remove
		out, err := r.Output([]string{"ps", "-a", "-q", "--filter", fmt.Sprintf("name=%s$", step.ContainerName())})
		if err != nil {
			return err
		}
		// Remove all found containers
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			if err := r.Exec([]string{"rm", scanner.Text()}); err != nil {
				return err
			}
		}
		return scanner.Err()
	}
}

// ContainerRunner returns a function to run the given step.
func (r *LocalRunner) ContainerRunner(step Step, network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Run container '%s'", step.ContainerName())
		}
		r.prefix = step.ColoredContainerName()
		r.stdout = step.Meta.Stdout
		r.stderr = step.Meta.Stderr
		return r.Exec(step.RunCommand(network))
	}
}

// NetworkCreator returns a function to create the given network.
func (r *LocalRunner) NetworkCreator(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Create network '%s'", network)
		}
		return r.Exec([]string{"network", "create", string(network)})
	}
}

// NetworkRemover returns a function to remove the given network.
func (r *LocalRunner) NetworkRemover(network Network) func() error {
	return func() error {
		if Verbose {
			log.Printf("Remove network '%s'", network)
		}
		return r.Exec([]string{"network", "rm", string(network)})
	}
}
