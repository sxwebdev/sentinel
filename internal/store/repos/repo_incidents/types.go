package repo_incidents

type StatsByServiceID struct {
	TotalIncidents      int64
	TotalDowntime       int64
	AvgDowntime         int64
	ResolvedIncidents   int64
	UnresolvedIncidents int64
	UptimePercentage30d float64
}

func (s StatsByMonitorIDRow) ToDomain() *StatsByServiceID {
	var totalDowntime int64
	if s.TotalDowntime != nil {
		totalDowntime = int64(*s.TotalDowntime)
	}

	var avgDowntime int64
	if s.AvgDowntime != nil {
		avgDowntime = int64(*s.AvgDowntime)
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
		UptimePercentage30d: s.UptimePercentage30d,
	}
}

type Stats struct {
	TotalIncidents      int64
	TotalDowntime       int64
	AvgDowntime         int64
	ResolvedIncidents   int64
	UnresolvedIncidents int64
}

func (s StatsRow) ToDomain() *Stats {
	var totalDowntime int64
	if s.TotalDowntime != nil {
		totalDowntime = int64(*s.TotalDowntime)
	}

	var avgDowntime int64
	if s.AvgDowntime != nil {
		avgDowntime = int64(*s.AvgDowntime)
	}

	var resolvedIncidents int64
	if s.ResolvedIncidents != nil {
		resolvedIncidents = int64(*s.ResolvedIncidents)
	}

	var unresolvedIncidents int64
	if s.UnresolvedIncidents != nil {
		unresolvedIncidents = int64(*s.UnresolvedIncidents)
	}

	return &Stats{
		TotalIncidents:      s.TotalIncidents,
		TotalDowntime:       totalDowntime,
		AvgDowntime:         avgDowntime,
		ResolvedIncidents:   resolvedIncidents,
		UnresolvedIncidents: unresolvedIncidents,
	}
}
