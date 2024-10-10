package server

import (
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
)

func Test_OrgRunnerMetricsExporter_getRunnersPerLabelMap(t *testing.T) {
	runners := []*github.Runner{
		{
			ID:     github.Int64(1),
			Name:   github.String("runner1"),
			Busy:   github.Bool(true),
			Status: github.String("online"),
			Labels: []*github.RunnerLabels{{Name: github.String("label1")}, {Name: github.String("label2")}},
		},
		{
			ID:     github.Int64(2),
			Name:   github.String("runner2"),
			Busy:   github.Bool(true),
			Status: github.String("online"),
			Labels: []*github.RunnerLabels{{Name: github.String("label1")}, {Name: github.String("label2")}},
		},
		{
			ID:     github.Int64(3),
			Name:   github.String("runner3"),
			Busy:   github.Bool(true),
			Status: github.String("online"),
			Labels: []*github.RunnerLabels{{Name: github.String("label3")}},
		},
		{
			ID:     github.Int64(4),
			Name:   github.String("runner4"),
			Busy:   github.Bool(false),
			Status: github.String("online"),
			Labels: []*github.RunnerLabels{{Name: github.String("label4")}},
		},
		{
			ID:     github.Int64(5),
			Name:   github.String("runner5"),
			Busy:   github.Bool(false),
			Status: github.String("offline"),
			Labels: []*github.RunnerLabels{{Name: github.String("label5")}},
		},
	}

	result := getRunnersPerLabelMap(runners)

	expected := map[string]map[string]int{
		"label1__label2": {"busy": 2},
		"label3":         {"busy": 1},
		"label4":         {"idle": 1},
	}

	assert.Equal(t, expected, result)
}
