package gantry

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestExamples(t *testing.T) {
	examples := []struct {
		dir   string
		def   string
		env   string
		cases []struct {
			key    string
			runner bool
			calls  int
			called int
		}
	}{
		{
			"diamond",
			"gantry.def.yml",
			"",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(a)", true, 1, 1},
				{"ImageExistenceChecker(b)", true, 1, 1},
				{"ImageExistenceChecker(c)", true, 1, 1},
				{"ImageExistenceChecker(d)", true, 1, 1},
				{"ContainerKiller(a)", true, 3, 3},
				{"ContainerKiller(b)", true, 3, 3},
				{"ContainerKiller(c)", true, 3, 3},
				{"ContainerKiller(d)", true, 3, 3},
				{"ContainerKiller(TempDirCleanUp)", true, 0, 0},
				{"ContainerRemover(a)", true, 4, 4},
				{"ContainerRemover(b)", true, 4, 4},
				{"ContainerRemover(c)", true, 4, 4},
				{"ContainerRemover(d)", true, 4, 4},
				{"ContainerRemover(TempDirCleanUp)", true, 0, 0},
				{"ContainerRunner(a,test)", true, 1, 1},
				{"ContainerRunner(b,test)", true, 1, 1},
				{"ContainerRunner(c,test)", true, 1, 1},
				{"ContainerRunner(d,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 0, 0},
			},
		},
		{
			"diamond_ignore_failure",
			"gantry.def.yml",
			"gantry.env.yml",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(a)", true, 1, 1},
				{"ImageExistenceChecker(b)", true, 1, 1},
				{"ImageExistenceChecker(c)", true, 1, 1},
				{"ImageExistenceChecker(d)", true, 1, 1},
				{"ContainerKiller(a)", true, 3, 3},
				{"ContainerKiller(b)", true, 3, 3},
				{"ContainerKiller(c)", true, 3, 3},
				{"ContainerKiller(d)", true, 3, 3},
				{"ContainerKiller(TempDirCleanUp)", true, 0, 0},
				{"ContainerRemover(a)", true, 4, 4},
				{"ContainerRemover(b)", true, 4, 4},
				{"ContainerRemover(c)", true, 4, 4},
				{"ContainerRemover(d)", true, 4, 4},
				{"ContainerRemover(TempDirCleanUp)", true, 0, 0},
				{"ContainerRunner(a,test)", true, 1, 1},
				{"ContainerRunner(b,test)", true, 1, 1},
				{"ContainerRunner(c,test)", true, 1, 1},
				{"ContainerRunner(d,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 0, 0},
			},
		},
		{
			"docker-compose",
			"docker-compose.yml",
			"",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(db)", true, 1, 1},
				{"ImageExistenceChecker(wordpress)", true, 1, 1},
				{"ContainerKiller(db)", true, 2, 2},
				{"ContainerKiller(wordpress)", true, 2, 2},
				{"ContainerKiller(TempDirCleanUp)", true, 0, 0},
				{"ContainerRemover(db)", true, 3, 3},
				{"ContainerRemover(wordpress)", true, 3, 3},
				{"ContainerRemover(TempDirCleanUp)", true, 0, 0},
				{"ContainerRunner(db,test)", true, 1, 1},
				{"ContainerRunner(wordpress,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 0, 0},
			},
		},
		{
			"partial_execution",
			"gantry.def.yml",
			"gantry.env.yml",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(service)", true, 1, 1},
				{"ImageExistenceChecker(wait_for_service)", true, 1, 1},
				{"ImageExistenceChecker(test_0)", true, 1, 1},
				{"ImageExistenceChecker(test_1)", true, 1, 1},
				{"ImageExistenceChecker(test_2)", true, 1, 1},
				{"ImageExistenceChecker(test_3)", true, 1, 1},
				{"ContainerKiller(service)", true, 3, 3},
				{"ContainerKiller(wait_for_service)", true, 3, 3},
				{"ContainerKiller(test_0)", true, 3, 3},
				{"ContainerKiller(test_1)", true, 3, 3},
				{"ContainerKiller(test_2)", true, 3, 3},
				{"ContainerKiller(test_3)", true, 3, 3},
				{"ContainerKiller(TempDirCleanUp)", true, 0, 0},
				{"ContainerRemover(service)", true, 4, 4},
				{"ContainerRemover(wait_for_service)", true, 4, 4},
				{"ContainerRemover(test_0)", true, 4, 4},
				{"ContainerRemover(test_1)", true, 4, 4},
				{"ContainerRemover(test_2)", true, 4, 4},
				{"ContainerRemover(test_3)", true, 4, 4},
				{"ContainerRemover(TempDirCleanUp)", true, 0, 0},
				{"ContainerRunner(service,test)", true, 1, 1},
				{"ContainerRunner(wait_for_service,test)", true, 1, 1},
				{"ContainerRunner(test_0,test)", true, 1, 1},
				{"ContainerRunner(test_1,test)", true, 1, 1},
				{"ContainerRunner(test_2,test)", true, 1, 1},
				{"ContainerRunner(test_3,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 0, 0},
			},
		},
		{
			"qlever_e2e",
			"gantry.def.yml",
			"gantry.env.yml",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(qlever)", true, 1, 1},
				{"ImageExistenceChecker(wait_for_qlever)", true, 0, 0},
				{"ImageExistenceChecker(download_input)", true, 1, 1},
				{"ImageExistenceChecker(unzip_input)", true, 1, 1},
				{"ImageExistenceChecker(build_index)", true, 1, 1},
				{"ImageExistenceChecker(run_queries)", true, 1, 1},
				{"ContainerKiller(qlever)", true, 3, 3},
				{"ContainerKiller(wait_for_qlever)", true, 3, 3},
				{"ContainerKiller(download_input)", true, 3, 3},
				{"ContainerKiller(unzip_input)", true, 3, 3},
				{"ContainerKiller(build_index)", true, 3, 3},
				{"ContainerKiller(run_queries)", true, 3, 3},
				{"ContainerKiller(TempDirCleanUp)", true, 1, 1},
				{"ContainerRemover(qlever)", true, 4, 4},
				{"ContainerRemover(wait_for_qlever)", true, 4, 4},
				{"ContainerRemover(download_input)", true, 4, 4},
				{"ContainerRemover(unzip_input)", true, 4, 4},
				{"ContainerRemover(build_index)", true, 4, 4},
				{"ContainerRemover(run_queries)", true, 4, 4},
				{"ContainerRemover(TempDirCleanUp)", true, 2, 2},
				{"ContainerRunner(qlever,test)", true, 1, 1},
				{"ContainerRunner(wait_for_qlever,test)", true, 1, 1},
				{"ContainerRunner(download_input,test)", true, 1, 1},
				{"ContainerRunner(unzip_input,test)", true, 1, 1},
				{"ContainerRunner(build_index,test)", true, 1, 1},
				{"ContainerRunner(run_queries,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 1, 1},
			},
		},
		{
			"selective_run",
			"gantry.def.yml",
			"gantry.env.yml",
			[]struct {
				key    string
				runner bool
				calls  int
				called int
			}{
				{"NetworkCreator(test)", true, 1, 1},
				{"ImageExistenceChecker(active_service)", true, 1, 1},
				{"ImageExistenceChecker(new_service)", true, 1, 1},
				{"ImageExistenceChecker(wait_for_new_service)", true, 1, 1},
				{"ImageExistenceChecker(pre_prepare_0)", true, 1, 1},
				{"ImageExistenceChecker(pre_prepare_1)", true, 1, 1},
				{"ImageExistenceChecker(prepare_new_service_version)", true, 1, 1},
				{"ImageExistenceChecker(test_new_service)", true, 1, 1},
				{"ImageExistenceChecker(move_data_to_active_service)", true, 1, 1},
				{"ContainerKiller(active_service)", true, 1, 1},
				{"ContainerKiller(new_service)", true, 3, 3},
				{"ContainerKiller(wait_for_new_service)", true, 3, 3},
				{"ContainerKiller(pre_prepare_0)", true, 3, 3},
				{"ContainerKiller(pre_prepare_1)", true, 3, 3},
				{"ContainerKiller(prepare_new_service_version)", true, 3, 3},
				{"ContainerKiller(test_new_service)", true, 3, 3},
				{"ContainerKiller(move_data_to_active_service)", true, 3, 3},
				{"ContainerKiller(TempDirCleanUp)", true, 0, 0},
				{"ContainerRemover(active_service)", true, 1, 1},
				{"ContainerRemover(new_service)", true, 4, 4},
				{"ContainerRemover(wait_for_new_service)", true, 4, 4},
				{"ContainerRemover(pre_prepare_0)", true, 4, 4},
				{"ContainerRemover(pre_prepare_1)", true, 4, 4},
				{"ContainerRemover(prepare_new_service_version)", true, 4, 4},
				{"ContainerRemover(test_new_service)", true, 4, 4},
				{"ContainerRemover(move_data_to_active_service)", true, 4, 4},
				{"ContainerRemover(TempDirCleanUp)", true, 0, 0},
				{"ContainerRunner(active_service,test)", true, 1, 1},
				{"ContainerRunner(new_service,test)", true, 1, 1},
				{"ContainerRunner(wait_for_new_service,test)", true, 1, 1},
				{"ContainerRunner(pre_prepare_0,test)", true, 1, 1},
				{"ContainerRunner(pre_prepare_1,test)", true, 1, 1},
				{"ContainerRunner(prepare_new_service_version,test)", true, 1, 1},
				{"ContainerRunner(test_new_service,test)", true, 1, 1},
				{"ContainerRunner(move_data_to_active_service,test)", true, 1, 1},
				{"ContainerRunner(TempDirCleanUp,test)", true, 0, 0},
			},
		},
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for _, example := range examples {
		if err := os.Chdir(filepath.Join(cwd, "examples", example.dir)); err != nil {
			log.Fatal(err)
		}
		p, err := NewPipeline(example.def, example.env, types.StringMap{}, types.StringSet{}, types.StringSet{})
		if err != nil {
			t.Errorf("Unexpected error creating pipeline for '%s': '%#v'", example.dir, err)
		}
		localRunner := NewNoopRunner(false)
		noopRunner := NewNoopRunner(false)
		p.localRunner = localRunner
		p.noopRunner = noopRunner
		p.Network = Network("test")

		if err := p.KillContainers(true); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.RemoveContainers(true); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.PullImages(false); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.BuildImages(false); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.CreateNetwork(); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.ExecuteSteps(); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}
		if err := p.CleanUp(syscall.SIGKILL); err != nil {
			t.Errorf("Unexpected error in '%s', got: '%#v', wanted 'nil'", example.dir, err)
		}

		for _, c := range example.cases {
			runner := noopRunner
			if c.runner {
				runner = localRunner
			}
			if v := runner.NumCalls(c.key); v != c.calls {
				t.Errorf("Incorrect NumCalls for '%s' in '%s', got: '%d', wanted '%d'", c.key, example.dir, v, c.calls)
			}
			if v := runner.NumCalled(c.key); v != c.called {
				t.Errorf("Incorrect NumCalled for '%s' in '%s', got: '%d', wanted '%d'", c.key, example.dir, v, c.called)
			}
		}
		if err := os.Chdir(cwd); err != nil {
			log.Fatal(err)
		}
	}
}

func TestExamplesAutodetectDefFiles(t *testing.T) {
	examples := []struct {
		dir string
	}{
		{
			"diamond",
		},
		{
			"diamond_ignore_failure",
		},
		{
			"docker-compose",
		},
		{
			"partial_execution",
		},
		{
			"qlever_e2e",
		},
		{
			"selective_run",
		},
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for i, example := range examples {
		if err := os.Chdir(filepath.Join(cwd, "examples", example.dir)); err != nil {
			log.Fatal(err)
		}
		if _, err := NewPipeline("", "", types.StringMap{}, types.StringSet{}, types.StringSet{}); err != nil {
			t.Errorf("Unexpected error creating pipeline for case '%d': '%s': '%#v'", i, example.dir, err)
		}
		if err := os.Chdir(cwd); err != nil {
			log.Fatal(err)
		}
	}
}
