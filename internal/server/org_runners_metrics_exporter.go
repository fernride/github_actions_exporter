package server

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type OrgRunnersMetricsExporter struct {
	GHClient *github.Client
	Logger   log.Logger
	Opts     Opts
}

func NewOrgRunnersMetricsExporter(logger log.Logger, opts Opts) *OrgRunnersMetricsExporter {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.GitHubAPIToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &OrgRunnersMetricsExporter{
		Logger:   logger,
		Opts:     opts,
		GHClient: client,
	}
}

func (c *OrgRunnersMetricsExporter) StartOrgRunnerApiPolling(ctx context.Context) error {
	if c.Opts.GitHubOrg == "" {
		return errors.New("github org not configured")
	}
	if c.Opts.GitHubAPIToken == "" {
		return errors.New("github token not configured")
	}

	ticker := time.NewTicker(time.Duration(c.Opts.OrgRunnersAPIPollSeconds) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.collectOrgRunnerApiPolling(ctx)
			case <-ctx.Done():
				_ = level.Info(c.Logger).Log("msg", "stopped polling for org runner metrics")
				return
			}
		}
	}()

	return nil
}

func concatenateLabels(labels []*github.RunnerLabels) string {
	var concatenatedLabels string
	for i, label := range labels {
		concatenatedLabels += label.GetName()
		if i < len(labels)-1 {
			concatenatedLabels += "__"
		}
	}
	return concatenatedLabels
}

func getRunnersPerLabelMap(runners []*github.Runner) map[string]map[string]int {
	runnersPerLabelMap := make(map[string]map[string]int)

	for _, runner := range runners {
		if runner.GetStatus() != "online" {
			continue
		}
		labels := concatenateLabels(runner.Labels)

		if _, exists := runnersPerLabelMap[labels]; !exists {
			runnersPerLabelMap[labels] = make(map[string]int)
		}
		if runner.GetBusy() {
			if _, exists := runnersPerLabelMap[labels]["busy"]; !exists {
				runnersPerLabelMap[labels]["busy"] = 0
			}
			runnersPerLabelMap[labels]["busy"]++
		} else {
			if _, exists := runnersPerLabelMap[labels]["idle"]; !exists {
				runnersPerLabelMap[labels]["idle"] = 0
			}
			runnersPerLabelMap[labels]["idle"]++
		}
	}

	return runnersPerLabelMap
}

func (c *OrgRunnersMetricsExporter) collectOrgRunnerApiPolling(ctx context.Context) {
	orgRunners, _, err := c.GHClient.Actions.ListOrganizationRunners(ctx, c.Opts.GitHubOrg, &github.ListRunnersOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	if err != nil {
		_ = c.Logger.Log("msg", "failed to retrieve org runner metrics for org", "org", c.Opts.GitHubOrg, "err", err)
		return
	}

	runnersPerLabelMap := getRunnersPerLabelMap(orgRunners.Runners)

	for labels, states := range runnersPerLabelMap {
		for state, count := range states {
			selfHostedRunnerCount.WithLabelValues(c.Opts.GitHubOrg, state, labels).Set(float64(count))
		}
	}
}
