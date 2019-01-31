package gantry_test

import (
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestColorStoreNextColor(t *testing.T) {
	var cases = []struct {
		store   *gantry.ColorStore
		results []int
	}{
		{gantry.NewColorStore([]int{42}), []int{42, 42, 42, 42, 42}},
		{gantry.NewColorStore([]int{21, 42}), []int{21, 42, 21, 42, 21}},
	}

	for _, c := range cases {
		for i, v := range c.results {
			color := c.store.NextColor()
			if color != v {
				t.Errorf("Incorrect color for index %d, got: %d, wanted %d", i, color, v)
			}
		}
	}
}
