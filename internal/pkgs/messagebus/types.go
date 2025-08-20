package messagebus

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkerStatus worker状态枚举
type WorkerStatus int32

const (
	WorkerStatusIdle       WorkerStatus = iota // 空闲等待
	WorkerStatusProcessing                     // 正在处理消息
	WorkerStatusStopping                       // 正在停止
	WorkerStatusStopped                        // 已停止
)

// String 返回状态的字符串表示
func (s WorkerStatus) String() string {
	switch s {
	case WorkerStatusIdle:
		return "idle"
	case WorkerStatusProcessing:
		return "processing"
	case WorkerStatusStopping:
		return "stopping"
	case WorkerStatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// WorkerInfo worker信息
type WorkerInfo struct {
	ID              int          `json:"id"`
	Status          WorkerStatus `json:"status"`
	CurrentTopic    string       `json:"current_topic,omitempty"`
	ProcessingStart *time.Time   `json:"processing_start,omitempty"`
	TotalProcessed  int64        `json:"total_processed"`
	LastProcessedAt *time.Time   `json:"last_processed_at,omitempty"`
}

// RuntimeStats 运行时状态统计
type RuntimeStats struct {
	WorkerCount       int            `json:"worker_count"`
	BufferSize        int            `json:"buffer_size"`
	QueueLength       int            `json:"queue_length"`
	IdleWorkers       int            `json:"idle_workers"`
	ProcessingWorkers int            `json:"processing_workers"`
	StoppingWorkers   int            `json:"stopping_workers"`
	StoppedWorkers    int            `json:"stopped_workers"`
	TotalTopics       int            `json:"total_topics"`
	TotalHandlers     int            `json:"total_handlers"`
	TopicHandlers     map[string]int `json:"topic_handlers"`
	Workers           []WorkerInfo   `json:"workers"`
}

// Handler 消息处理函数类型
type Handler func(context.Context, interface{})

// Message 消息结构
type Message struct {
	Topic   string
	Payload interface{}
	Ctx     context.Context
	Done    chan struct{} // 用于同步发布的完成通知
}

// MessageBus 消息总线接口
type MessageBus interface {
	Subscribe(topic string, handler Handler)
	Publish(ctx context.Context, topic string, payload interface{}) error
	PublishSync(ctx context.Context, topic string, payload interface{}) error
	Close()
	QueueLength() int
	WorkerCount() int
	GetRuntimeStats() RuntimeStats
	GetWorkerInfo(workerID int) (WorkerInfo, bool)
	GetAllWorkerInfo() []WorkerInfo
}

// ConcurrentMessageBus 支持并发的消息总线实现
type ConcurrentMessageBus struct {
	subscribers map[string][]Handler
	queue       chan Message
	workers     []*worker
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	workerCount int
	bufferSize  int
	logger      *zap.Logger
}

// Config 配置选项
type Config struct {
	WorkerCount int // 并发worker数量，1表示单线程
	BufferSize  int // 队列缓冲区大小
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		WorkerCount: 1, // 默认单线程
		BufferSize:  8, // 默认缓冲区大小
	}
}
