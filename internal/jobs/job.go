package jobs

import (
	"context"

	"github.com/pkg/errors"
)

type Job interface {
	Start(ctx context.Context) error
	Stop()
}

var (
	ErrJobRunning = errors.New("job is running")
)
