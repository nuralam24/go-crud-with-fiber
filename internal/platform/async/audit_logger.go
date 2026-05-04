package async

import (
	"log/slog"
	"sync"
)

type AuditEvent struct {
	Action string
	Actor  string
	Target string
}

type AuditLogger struct {
	ch      chan AuditEvent
	workers int
	logger  *slog.Logger
	wg      sync.WaitGroup
	mu      sync.Mutex
	stopped bool
}

func NewAuditLogger(bufferSize int, workers int, logger *slog.Logger) *AuditLogger {
	return &AuditLogger{
		ch:      make(chan AuditEvent, bufferSize),
		workers: workers,
		logger:  logger,
	}
}

// Start launches worker goroutines. Shutdown is driven solely by Stop (closes the channel);
// workers drain until the channel is closed, so buffered events are not abandoned on shutdown.
func (a *AuditLogger) Start() {
	a.wg.Add(a.workers)
	for range a.workers {
		go func() {
			defer a.wg.Done()
			for event := range a.ch {
				a.logger.Info("audit-event", "action", event.Action, "actor", event.Actor, "target", event.Target)
			}
		}()
	}
}

func (a *AuditLogger) Stop() {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return
	}
	a.stopped = true
	close(a.ch)
	a.mu.Unlock()
	a.wg.Wait()
}

func (a *AuditLogger) Publish(event AuditEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.stopped {
		return
	}
	select {
	case a.ch <- event:
	default:
		// Drop when under pressure to keep request path ultra-fast.
	}
}
