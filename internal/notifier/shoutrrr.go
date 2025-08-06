package notifier

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/tkcrm/mx/logger"
)

// notificationRequest represents a single notification request
type notificationRequest struct {
	Message   string
	CreatedAt time.Time
}

type Notifier struct {
	logger    logger.Logger
	urls      []string
	queue     chan *notificationRequest
	done      chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	isStarted bool
	name      string
}

// New creates a new Notifier instance as a service
func New(l logger.Logger, urls []string) (*Notifier, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no notification URLs provided")
	}

	notifier := &Notifier{
		logger: l,
		urls:   urls,
		queue:  make(chan *notificationRequest, 500), // buffer for 500 notifications
		done:   make(chan struct{}),
		name:   "NotificationService",
	}

	return notifier, nil
}

// Name returns the service name
func (s *Notifier) Name() string { return s.name }

// Start starts the notification service
func (s *Notifier) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStarted {
		return fmt.Errorf("notification service already started")
	}

	s.isStarted = true
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.processQueue()
	}()

	return nil
}

// Stop stops the notification service
func (s *Notifier) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.isStarted {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	// Close the done channel to signal stop
	close(s.done)

	// Wait for completion with timeout from context
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed
	case <-ctx.Done():
		// Timeout - force termination
		return fmt.Errorf("notification service stop timeout: %w", ctx.Err())
	}

	s.mu.Lock()
	s.isStarted = false
	s.mu.Unlock()

	return nil
}

// processQueue processes the notification queue with 500ms delay
func (s *Notifier) processQueue() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			// Process remaining notifications in queue before exit
			for {
				select {
				case req := <-s.queue:
					s.processNotification(req)
				default:
					return
				}
			}

		case <-ticker.C:
			// Check queue every 500ms
			select {
			case req := <-s.queue:
				s.processNotification(req)
			default:
				// Queue is empty, continue
			}
		}
	}
}

// processNotification processes a single notification with simple retry
func (s *Notifier) processNotification(req *notificationRequest) {
	err := s.sendMessageSync(req.Message)

	if err != nil {
		s.logger.Errorf("failed to send notification: %v", err)

		// Put back in queue for retry
		s.queue <- req
	} else {
		s.logger.Info("notification sent successfully")
	}
}

// SendAlert sends an alert notification when a service goes down
func (s *Notifier) SendAlert(service *storage.Service, incident *storage.Incident) error {
	s.mu.RLock()
	if !s.isStarted {
		s.mu.RUnlock()
		return fmt.Errorf("notification service is not started")
	}
	s.mu.RUnlock()

	message := s.formatAlertMessage(service, incident)
	return s.enqueueMessage(message)
}

// SendRecovery sends a recovery notification when a service comes back up
func (s *Notifier) SendRecovery(service *storage.Service, incident *storage.Incident) error {
	s.mu.RLock()
	if !s.isStarted {
		s.mu.RUnlock()
		return fmt.Errorf("notification service is not started")
	}
	s.mu.RUnlock()

	message := s.formatRecoveryMessage(service, incident)
	return s.enqueueMessage(message)
}

// enqueueMessage adds a message to the queue for sending
func (s *Notifier) enqueueMessage(message string) error {
	req := &notificationRequest{
		Message:   message,
		CreatedAt: time.Now(),
	}

	select {
	case s.queue <- req:
		return nil
	default:
		return fmt.Errorf("notification queue is full")
	}
}

// sendMessageSync sends a message to all configured providers synchronously
// If one provider fails, it continues with others and returns partial errors
func (s *Notifier) sendMessageSync(message string) error {
	// Send to all providers concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(s.urls))

	for i, url := range s.urls {
		wg.Add(1)
		go func(index int, providerURL string) {
			defer wg.Done()

			// Create individual sender for this provider
			sender, err := shoutrrr.CreateSender(providerURL)
			if err != nil {
				errors <- fmt.Errorf("provider %d: failed to create sender: %w", index, err)
				return
			}

			providerName, _, err := sender.ExtractServiceName(providerURL)
			if err != nil {
				errors <- fmt.Errorf("provider %d: failed to extract service name: %w", index, err)
				return
			}

			// Send message with timeout
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer timeoutCancel()

			done := make(chan []error, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						errors <- fmt.Errorf("provider %s: panic during send: %v", providerName, r)
					}
				}()
				done <- sender.Send(message, nil)
			}()

			select {
			case errs := <-done:
				// Check if there are any actual errors (not nil)
				hasErrors := false
				for _, err := range errs {
					if err != nil {
						hasErrors = true
						break
					}
				}

				if hasErrors {
					errors <- fmt.Errorf("provider %s: failed to send: %v", providerName, errs)
				}
			case <-timeoutCtx.Done():
				errors <- fmt.Errorf("provider %s: timeout", providerName)
				// Note: The goroutine with sender.Send() might still be running,
				// but we can't force-kill it. This is a limitation of the shoutrrr library.
			}
		}(i, url)
	}

	wg.Wait()
	close(errors)

	// Collect all errors
	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	// If all providers failed, return an error
	if len(allErrors) == len(s.urls) {
		return fmt.Errorf("all notification providers failed: %v", allErrors)
	}

	// If some providers failed, return partial error
	if len(allErrors) > 0 {
		return fmt.Errorf("some notification providers failed (%d/%d): %v", len(allErrors), len(s.urls), allErrors)
	}

	return nil
}

// formatAlertMessage formats an alert message
func (s *Notifier) formatAlertMessage(service *storage.Service, incident *storage.Incident) string {
	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🔴 [ALERT] %s is DOWN\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Error: %s\n"+
			"• Started: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
		incident.Error,
		incident.StartTime.Format("2006-01-02 15:04:05"),
		incident.ID,
	)
}

// formatRecoveryMessage formats a recovery message
func (s *Notifier) formatRecoveryMessage(service *storage.Service, incident *storage.Incident) string {
	var duration string
	if incident.Duration != nil {
		duration = formatDuration(*incident.Duration)
	} else {
		duration = formatDuration(time.Since(incident.StartTime))
	}

	var endTime string
	if incident.EndTime != nil {
		endTime = incident.EndTime.Format("2006-01-02 15:04:05")
	} else {
		endTime = time.Now().Format("2006-01-02 15:04:05")
	}

	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🟢 [RECOVERY] %s is UP\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Downtime: %s\n"+
			"• Recovered: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
		duration,
		endTime,
		incident.ID,
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
