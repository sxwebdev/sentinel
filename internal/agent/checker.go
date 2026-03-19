package agent

import (
	"context"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
)

type checkResult struct {
	serviceID            string
	status               models.ServiceStatus
	checkedAt            time.Time
	responseMessage      string
	responseErrorMessage string
	responseTime         time.Duration
}

func (s *Agent) onSuccess(ctx context.Context, serviceID string, responseTime time.Duration) error {
	s.checkResultsCh <- checkResult{
		serviceID:    serviceID,
		status:       models.ServiceStatusUp,
		checkedAt:    time.Now(),
		responseTime: responseTime,
	}
	return nil
}

func (s *Agent) onFailure(ctx context.Context, serviceID string, checkErr error, responseTime time.Duration) error {
	s.checkResultsCh <- checkResult{
		serviceID:            serviceID,
		status:               models.ServiceStatusDown,
		checkedAt:            time.Now(),
		responseErrorMessage: checkErr.Error(),
		responseTime:         responseTime,
	}
	return nil
}
