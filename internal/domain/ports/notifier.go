// internal/domain/ports/notifier.go
package ports

import "frauddetection/internal/domain"

// FraudNotifier defines the interface for sending fraud alerts to external systems.
type FraudNotifier interface {
	Notify(event domain.FraudAlert) error
}
