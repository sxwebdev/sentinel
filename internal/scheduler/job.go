package scheduler

import (
	"context"
	"sync/atomic"
	"time"
)

// job represents a scheduled monitoring job for a service
type job struct {
	serviceID   string
	serviceName string
	interval    time.Duration
	timeout     time.Duration
	retries     int64
	ticker      *time.Ticker
	stopChan    chan struct{}
	inProgress  atomic.Bool
	// Context and cancel function for canceling ongoing checks
	checkCtx    context.Context
	checkCancel context.CancelFunc
}
