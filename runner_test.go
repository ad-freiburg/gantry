package gantry_test

import (
	"fmt"
	"testing"

	"github.com/ad-freiburg/gantry"
)

const stepName string = "step"
const networkName string = "network"

func TestNoopRunnerImageBuilder(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageBuilder(%s,%t)", step.Name, false)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ImageBuilder(step, false)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerImageBuilderForcePull(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageBuilder(%s,%t)", step.Name, true)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ImageBuilder(step, true)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerImagePuller(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImagePuller(%s)", step.Name)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ImagePuller(step)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerImageExistenceChecker(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ImageExistenceChecker(%s)", step.Name)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ImageExistenceChecker(step)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerContainerKiller(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ContainerKiller(%s)", step.Name)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ContainerKiller(step)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	num, err := f()
	if err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if num != 0 {
		t.Errorf("Incorrect number of killed containers, got: '%#v', wanted '0'", num)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerContainerRemover(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	key := fmt.Sprintf("ContainerRemover(%s)", step.Name)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ContainerRemover(step)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerContainerRunner(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	step := gantry.Step{}
	step.Name = stepName
	network := gantry.Network(networkName)
	key := fmt.Sprintf("ContainerRunner(%s,%s)", step.Name, network)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.ContainerRunner(step, network)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerNetworkCreator(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	network := gantry.Network(networkName)
	key := fmt.Sprintf("NetworkCreator(%s)", network)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.NetworkCreator(network)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}

func TestNoopRunnerNetworkRemover(t *testing.T) {
	runner := gantry.NewNoopRunner(true)
	network := gantry.Network(networkName)
	key := fmt.Sprintf("NetworkRemover(%s)", network)
	if c := runner.NumCalls(key); c != 0 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 0)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	f := runner.NetworkRemover(network)
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 0 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 0)
	}

	if err := f(); err != nil {
		t.Errorf("Unexpected error, got: '%#v', wanted 'nil'", err)
	}
	if c := runner.NumCalls(key); c != 1 {
		t.Errorf("Incorrect NumCalls for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
	if c := runner.NumCalled(key); c != 1 {
		t.Errorf("Incorrect NumCalled for '%s', got: '%d', wanted '%d'", key, c, 1)
	}
}
