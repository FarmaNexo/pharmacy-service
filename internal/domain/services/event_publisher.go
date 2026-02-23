// internal/domain/services/event_publisher.go
package services

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/events"
)

type EventPublisher interface {
	Publish(ctx context.Context, event events.PharmacyEvent) error
}
