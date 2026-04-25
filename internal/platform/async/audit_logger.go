package async

import (
	"context"
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
}

func NewAuditLogger(bufferSize int, workers int, logger *slog.Logger) *AuditLogger {
	return &AuditLogger{
		ch:      make(chan AuditEvent, bufferSize),
		workers: workers,
		logger:  logger,
	}
}

func (a *AuditLogger) Start(ctx context.Context) {
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-a.ch:
					if !ok {
						return
					}
					a.logger.Info("audit-event", "action", event.Action, "actor", event.Actor, "target", event.Target)
				}
			}
		}()
	}
}

func (a *AuditLogger) Stop() {
	close(a.ch)
	a.wg.Wait()
}

func (a *AuditLogger) Publish(event AuditEvent) {
	select {
	case a.ch <- event:
	default:
		// Drop when under pressure to keep request path ultra-fast.
	}
}
