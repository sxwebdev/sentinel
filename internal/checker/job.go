package checker

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
)

// job represents a scheduled monitoring job for a service
type job struct {
	serviceID       string
	serviceName     string
	serviceConfig   map[string]any
	serviceProtocol models.ServiceProtocolType
	interval        time.Duration
	timeout         time.Duration
	retries         int64
	ticker          *time.Ticker
	stopChan        chan struct{}
	inProgress      atomic.Bool
	// Context and cancel function for canceling ongoing checks
	checkCtx    context.Context
	checkCancel context.CancelFunc
}
