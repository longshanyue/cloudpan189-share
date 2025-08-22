package eventbus

import (
	"fmt"
	"sync/atomic"
)

var globalID int64

func generateID(counter int64) string {
	id := atomic.AddInt64(&globalID, 1)
	return fmt.Sprintf("sub_%d_%d", counter, id)
}
