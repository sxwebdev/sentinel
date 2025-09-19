package repo_incidents

import "time"

type StatsByServiceID struct {
	TotalIncidents      int64
	TotalDowntime       time.Duration
	AvgDowntime         time.Duration
	ResolvedIncidents   int64
	UnresolvedIncidents int64
}

func (s StatsByServiceIDRow) ToDomain() *StatsByServiceID {
	var totalDowntime time.Duration
	if s.TotalDowntime != nil {
		totalDowntime = time.Duration(*s.TotalDowntime)
	}

	var avgDowntime time.Duration
	if s.AvgDowntime != nil {
		avgDowntime = time.Duration(*s.AvgDowntime)
	}

	var resolvedIncidents int64
	if s.ResolvedIncidents != nil {
		resolvedIncidents = int64(*s.ResolvedIncidents)
	}

	var unresolvedIncidents int64
	if s.UnresolvedIncidents != nil {
		unresolvedIncidents = int64(*s.UnresolvedIncidents)
	}

	return &StatsByServiceID{
		TotalIncidents:      s.TotalIncidents,
		TotalDowntime:       totalDowntime,
		AvgDowntime:         avgDowntime,
		ResolvedIncidents:   resolvedIncidents,
		UnresolvedIncidents: unresolvedIncidents,
	}
}
