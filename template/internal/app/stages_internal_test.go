package app

import "testing"

// The stage table ascends in the process's dependency order. A stage a
// service adds between two rows, or a library constant it names, joins
// the list here so a release that moved it out of order fails instead of
// reordering startup silently.
func TestStages_Ascend(t *testing.T) {
	stages := []struct {
		name  string
		stage int
	}{
		{"infrastructure", stageInfrastructure},
		{"root", stageRoot},
	}
	for i := 1; i < len(stages); i++ {
		if prev, cur := stages[i-1], stages[i]; cur.stage <= prev.stage {
			t.Errorf("stage %s (%d) does not follow %s (%d)", cur.name, cur.stage, prev.name, prev.stage)
		}
	}
}
