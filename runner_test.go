package gantry_test

import (
	"fmt"
	"testing"

	"github.com/ad-freiburg/gantry"
)

const stepName string = "step"
const networkName string = "network"

func checkCallsAndCalled(t *testing.T, runner *gantry.NoopRunner, key string, calls int, called int) {
	if c := runner.NumCalls(key); c != calls {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, calls)
	}
	if c := runner.NumCalled(key); c != called {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, called)
	}
}

func TestNoopRunnerImageBuilder(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageBuilder(%s,%t)", step.Name, false)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ImageBuilder(step, false)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerImageBuilderForcePull(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageBuilder(%s,%t)", step.Name, true)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ImageBuilder(step, true)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerImagePuller(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImagePuller(%s)", step.Name)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ImagePuller(step)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerImageExistenceChecker(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageExistenceChecker(%s)", step.Name)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ImageExistenceChecker(step)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerContainerKiller(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ContainerKiller(%s)", step.Name)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ContainerKiller(step)
	checkCallsAndCalled(t, runner, key, 1, 0)

	num, err := f()
	if err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if num != 0 {
		t.Errorf("Incorrect number of killed containers, got: '%#v', wanted '0'", num)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerContainerRemover(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ContainerRemover(%s)", step.Name)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ContainerRemover(step)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerContainerRunner(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	network := gantry.Network(networkName)
	key := fmt.Sprintf("ContainerRunner(%s,%s)", step.Name, network)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.ContainerRunner(step, network)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerNetworkCreator(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	network := gantry.Network(networkName)
	key := fmt.Sprintf("NetworkCreator(%s)", network)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.NetworkCreator(network)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}

func TestNoopRunnerNetworkRemover(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	network := gantry.Network(networkName)
	key := fmt.Sprintf("NetworkRemover(%s)", network)
	checkCallsAndCalled(t, runner, key, 0, 0)

	f := runner.NetworkRemover(network)
	checkCallsAndCalled(t, runner, key, 1, 0)

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	checkCallsAndCalled(t, runner, key, 1, 1)
}
