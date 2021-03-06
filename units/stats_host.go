package units

import (
	"context"
	"fmt"

	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/dependency"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/logging"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

const hostStatsCollectorJobName = "host-stats-collector"

func init() {
	registry.AddJobType(hostStatsCollectorJobName,
		func() amboy.Job { return makeHostStatsCollector() })
}

type hostStatsCollector struct {
	job.Base `bson:"job_base" json:"job_base" yaml:"job_base"`
	logger   grip.Journaler
}

// NewHostStatsCollector logs statistics about host utilization per
// distro to the default grip logger.
func NewHostStatsCollector(id string) amboy.Job {
	j := makeHostStatsCollector()
	j.SetID(fmt.Sprintf("%s-%s", hostStatsCollectorJobName, id))
	j.SetPriority(-1)

	return j
}

func makeHostStatsCollector() *hostStatsCollector {
	j := &hostStatsCollector{
		logger: logging.MakeGrip(grip.GetSender()),
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    hostStatsCollectorJobName,
				Version: 0,
			},
		},
	}
	j.SetDependency(dependency.NewAlways())
	return j
}

func (j *hostStatsCollector) Run(_ context.Context) {
	defer j.MarkComplete()

	if j.logger == nil {
		j.logger = logging.MakeGrip(grip.GetSender())
	}

	j.AddError(j.statsByDistro())
	j.AddError(j.statsByProvider())
}

func (j *hostStatsCollector) statsByDistro() error {
	hosts, err := host.GetStatsByDistro()
	if err != nil {
		return errors.Wrap(err, "problem getting stats by distro")
	}

	tasks := 0
	count := 0
	excess := 0
	for _, h := range hosts {
		count += h.Count
		tasks += h.NumTasks

		overage := -1 * (h.MaxHosts - h.Count)
		if overage > 0 {
			excess += overage
		}
	}

	j.logger.Info(message.Fields{
		"report":        "host stats by distro",
		"hosts_total":   count,
		"running_tasks": tasks,
		"data":          hosts,
	})

	j.logger.WarningWhen(excess > 0, message.Fields{
		"report": "maxHosts exceeded",
		"data":   hosts.MaxHostsExceeded(),
		"total":  excess,
	})

	return nil
}

func (j *hostStatsCollector) statsByProvider() error {
	providers, err := host.GetProviderCounts()
	if err != nil {
		return errors.Wrap(err, "problem getting stats by provider")
	}

	j.logger.Info(message.Fields{
		"report": "host stats by provider",
		// or we could make providers a map of provider names
		// (string) to counts, by calling .Map() on the providers value.
		"providers": providers,
	})

	return nil
}
