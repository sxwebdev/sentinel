package repo_service_states

type ServicesStats struct {
	TotalServices   int64
	ServicesUp      int64
	ServicesDown    int64
	ServicesUnknown int64
	AvgResponseTime int64
	TotalChecks     int64
}

func (s StatsRow) ToDomain() *ServicesStats {
	var servicesUp int64
	if s.ServicesUp != nil {
		servicesUp = int64(*s.ServicesUp)
	}

	var servicesDown int64
	if s.ServicesDown != nil {
		servicesDown = int64(*s.ServicesDown)
	}

	var servicesUnknown int64
	if s.ServicesUnknown != nil {
		servicesUnknown = int64(*s.ServicesUnknown)
	}

	var avgResponseTime int64
	if s.AvgResponseTime != nil {
		avgResponseTime = int64(*s.AvgResponseTime)
	}

	var totalChecks int64
	if s.TotalChecks != nil {
		totalChecks = int64(*s.TotalChecks)
	}

	dto := &ServicesStats{
		TotalServices:   s.TotalServices,
		ServicesUp:      servicesUp,
		ServicesDown:    servicesDown,
		ServicesUnknown: servicesUnknown,
		AvgResponseTime: avgResponseTime,
		TotalChecks:     totalChecks,
	}

	return dto
}
