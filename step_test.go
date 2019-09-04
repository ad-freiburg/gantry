package gantry_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
)

func TestStepCheck(t *testing.T) {
	cases := []struct {
		step gantry.Step
		err  bool
	}{
		{gantry.Step{Service: gantry.Service{Name: "a"}}, true},
		{gantry.Step{Service: gantry.Service{Name: "a", Image: "alpine"}}, false},
		{gantry.Step{Service: gantry.Service{Name: "a", Image: "alpine", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}}, false},
		{gantry.Step{Service: gantry.Service{Name: "a", Image: "alpine", Restart: "always", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}}, true},
		{gantry.Step{Service: gantry.Service{Name: "a", Image: "alpine", Restart: "always", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeService}}}, false},
	}

	for i, c := range cases {
		err := c.step.Check()
		if err != nil && !c.err {
			t.Errorf("Unexpected error for case '%d': '%#v', got: '%#v'", i, c.step, err)
		}
		if err == nil && c.err {
			t.Errorf("Expected error for case '%d': '%#v', got: 'nil'", i, c.step)
		}
	}
}

func TestStepDependencies(t *testing.T) {
	cases := []struct {
		step   gantry.Step
		result types.StringSet
	}{
		{
			gantry.Step{Service: gantry.Service{Name: "a"}},
			types.StringSet{},
		},
		{
			gantry.Step{Service: gantry.Service{Name: "b"}, After: map[string]bool{"a": true}},
			types.StringSet{"a": true},
		},
		{
			gantry.Step{Service: gantry.Service{Name: "b", DependsOn: map[string]bool{"a": true}}},
			types.StringSet{"a": true},
		},
		{
			gantry.Step{Service: gantry.Service{Name: "d", DependsOn: map[string]bool{"c": true}}, After: map[string]bool{"b": true}},
			types.StringSet{"b": true, "c": true},
		},
	}

	for i, c := range cases {
		r := c.step.Dependencies()
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for case '%d': '%v', got: '%#v', wanted '%#v'", i, c.step, r, c.result)
		}
	}
}

func TestStepImageName(t *testing.T) {
	cases := []struct {
		step   gantry.Step
		result string
	}{
		{
			gantry.Step{Service: gantry.Service{Name: "a Step"}},
			"a_step",
		},
		{
			gantry.Step{Service: gantry.Service{Image: "b"}},
			"b",
		},
		{
			gantry.Step{Service: gantry.Service{Name: "c Step", Image: "c"}},
			"c",
		},
	}

	for _, c := range cases {
		r := c.step.ImageName()
		if r != c.result {
			t.Errorf("Incorrect result for '%v', got: '%s', wanted '%s'", c.step, r, c.result)
		}
	}
}

func TestStepRawContainerName(t *testing.T) {
	cases := []struct {
		step   gantry.Step
		result string
	}{
		{
			gantry.Step{Service: gantry.Service{Name: "a Step"}},
			"a_step",
		},
		{
			gantry.Step{Service: gantry.Service{Image: "b"}},
			"",
		},
		{
			gantry.Step{Service: gantry.Service{Name: "c Step", Image: "c"}},
			"c_step",
		},
	}

	for _, c := range cases {
		r := c.step.RawContainerName()
		if r != c.result {
			t.Errorf("Incorrect result for '%v', got: '%s', wanted '%s'", c.step, r, c.result)
		}
	}
}

func TestStepContainerName(t *testing.T) {
	gantry.ProjectName = "P"

	cases := []struct {
		step   gantry.Step
		result string
	}{
		{
			gantry.Step{Service: gantry.Service{Name: "a Step"}},
			"P_a_step",
		},
		{
			gantry.Step{Service: gantry.Service{Image: "b"}},
			"P_",
		},
		{
			gantry.Step{Service: gantry.Service{Name: "c Step", Image: "c"}},
			"P_c_step",
		},
	}

	for _, c := range cases {
		r := c.step.ContainerName()
		if r != c.result {
			t.Errorf("Incorrect result for '%v', got: '%s', wanted '%s'", c.step, r, c.result)
		}
	}
}

func TestStepBuildCommand(t *testing.T) {
	bar := "Bar"
	cases := []struct {
		step   gantry.Step
		pull   bool
		result []string
	}{
		{
			gantry.Step{Service: gantry.Service{Image: "img"}},
			false,
			[]string{"build", "--tag", "img", "."},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img"}},
			true,
			[]string{"build", "--tag", "img", "--pull", "."},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", BuildInfo: gantry.BuildInfo{Dockerfile: "file"}}},
			false,
			[]string{"build", "--tag", "img", "--file", "file", "."},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", BuildInfo: gantry.BuildInfo{Context: "./context"}}},
			false,
			[]string{"build", "--tag", "img", "./context"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", BuildInfo: gantry.BuildInfo{Args: map[string]*string{"Foo": &bar}}}},
			false,
			[]string{"build", "--tag", "img", "--build-arg", "Foo=Bar", "."},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", BuildInfo: gantry.BuildInfo{Args: map[string]*string{"USER": nil}}}},
			false,
			[]string{"build", "--tag", "img", "--build-arg", fmt.Sprintf("USER=%s", os.Getenv("USER")), "."},
		},
	}

	for _, c := range cases {
		r := c.step.BuildCommand(c.pull)
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v',pull:%t , got: '%v', wanted '%v'", c.step, c.pull, r, c.result)
		}
	}
}

func TestStepRunCommand(t *testing.T) {
	bar := "Bar"
	cases := []struct {
		step    gantry.Step
		network gantry.Network
		result  []string
	}{
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "i", Name: "n", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeService}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_n", "--network", "dummy", "--network-alias", "n", "--network-alias", "T_n", "-d", "i"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Ports: []string{"8080:5000"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "-p", "8080:5000", "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Environment: map[string]*string{"Foo": &bar}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "-e", "Foo=Bar", "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Environment: map[string]*string{"USER": nil}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "-e", fmt.Sprintf("USER=%s", os.Getenv("USER")), "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Volumes: []string{"/tmp:/tmp"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "-v", "/tmp:/tmp", "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Command: types.StringOrStringSlice{"Do", "nothing"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "img", "Do", "nothing"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Command: types.StringOrStringSlice{"Do nothing"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "img", "Do", "nothing"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Entrypoint: types.StringOrStringSlice{"Do", "nothing"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "--entrypoint", "Do", "img", "nothing"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Entrypoint: types.StringOrStringSlice{"Do nothing"}, Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "--entrypoint", "Do", "img", "nothing"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Restart: "never", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeStep}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "--rm", "--restart", "never", "img"},
		},
		{
			gantry.Step{Service: gantry.Service{Image: "img", Name: "name", Restart: "unless-stopped", Meta: gantry.ServiceMeta{Type: gantry.ServiceTypeService}}},
			gantry.Network("dummy"),
			[]string{"run", "--name", "T_name", "--network", "dummy", "--network-alias", "name", "--network-alias", "T_name", "-d", "--restart", "unless-stopped", "img"},
		},
	}

	gantry.ProjectName = "T"
	for i, c := range cases {
		r := c.step.RunCommand(c.network)
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for test '%d': '%v',network:'%s' , got: '%v', wanted '%v'", i, c.step, c.network, r, c.result)
		}
	}
}

func TestStepPullCommand(t *testing.T) {
	cases := []struct {
		step   gantry.Step
		result []string
	}{
		{gantry.Step{Service: gantry.Service{Image: "img"}}, []string{"pull", "img"}},
	}

	for _, c := range cases {
		r := c.step.PullCommand()
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v', got: '%v', wanted '%v'", c.step, r, c.result)
		}
	}
}
