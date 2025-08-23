// eventbus/errors.go
package eventbus

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEventBusClosed     = errors.New("event bus is closed")
	ErrChannelFull        = errors.New("event channel is full")
	ErrNoSubscribers      = errors.New("no subscribers for topic")
	ErrSubscriptionClosed = errors.New("subscription is closed")
)

// PublishError 发布错误，包含详细信息
type PublishError struct {
	Topic        string
	SuccessCount int
	FailedCount  int
	LastError    error
}

func (pe *PublishError) Error() string {
	return fmt.Sprintf("publish to topic '%s': %d succeeded, %d failed, last error: %v",
		pe.Topic, pe.SuccessCount, pe.FailedCount, pe.LastError)
}

func (pe *PublishError) Unwrap() error {
	return pe.LastError
}

// MultiError 多个错误
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}

	return fmt.Sprintf("multiple errors: [%s]", strings.Join(msgs, ", "))
}

func (e *MultiError) Unwrap() []error {
	return e.Errors
}
