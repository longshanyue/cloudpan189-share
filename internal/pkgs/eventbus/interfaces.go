package eventbus

import (
	"context"
	"time"
)

// Handler 事件处理函数
type Handler func(ctx context.Context, data interface{}) error

// Subscription 订阅接口
type Subscription interface {
	ID() string
	Topic() string
	Close()
}

// TaskInfo 任务信息
type TaskInfo struct {
	ID        string      `json:"id"`
	Topic     string      `json:"topic"`
	Status    string      `json:"status"` // "pending", "running", "completed"
	StartTime time.Time   `json:"startTime"`
	Data      interface{} `json:"data,omitempty"`
}

// BusStats 总线统计信息
type BusStats struct {
	RunningCount     int   `json:"runningCount"`
	PendingCount     int   `json:"pendingCount"`
	CompletedCount   int64 `json:"completedCount"`
	QueueLength      int   `json:"queueLength"`
	ActiveWorkers    int   `json:"activeWorkers"`
	TotalSubscribers int   `json:"totalSubscribers"`
}

// EventBus 事件总线接口
type EventBus interface {
	Subscribe(topic string, handler Handler) Subscription
	Unsubscribe(sub Subscription)
	Publish(ctx context.Context, topic string, data interface{}) error
	PublishSync(ctx context.Context, topic string, data interface{}) error
	Close()

	// GetRunningTasks 任务状态查询接口
	GetRunningTasks() []TaskInfo
	GetPendingTasks() []TaskInfo
	GetStats() BusStats
}
