package eventbus

import (
	"context"
	"time"
)

// Event 事件
type Event struct {
	ID        string // 事件唯一标识
	Topic     string
	Data      interface{}
	Context   context.Context
	Handler   Handler
	Result    chan error // 同步发布时使用
	Status    string     // 事件状态: "pending", "running", "completed"
	StartTime time.Time  // 创建时间
}
