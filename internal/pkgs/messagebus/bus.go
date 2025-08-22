package messagebus

import (
	"context"

	"github.com/pkg/errors"

	"go.uber.org/zap"
)

// New 创建新的消息总线
func New(config Config, logger *zap.Logger) MessageBus {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 1
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 8
	}

	ctx, cancel := context.WithCancel(context.Background())

	bus := &ConcurrentMessageBus{
		subscribers: make(map[string][]Handler),
		queue:       make(chan Message, config.BufferSize),
		ctx:         ctx,
		cancel:      cancel,
		workerCount: config.WorkerCount,
		bufferSize:  config.BufferSize,
		logger:      logger.With(zap.String("component", "message_bus")),
	}

	// 创建并启动workers
	bus.workers = make([]*worker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		bus.workers[i] = newWorker(i+1, bus, &bus.wg)
		bus.workers[i].start()
	}

	bus.logger.Info("消息总线已创建",
		zap.Int("worker_count", config.WorkerCount),
		zap.Int("buffer_size", config.BufferSize))

	return bus
}

// Subscribe 订阅主题
func (b *ConcurrentMessageBus) Subscribe(topic string, handler Handler) {
	if handler == nil {
		b.logger.Warn("Nil handler for topic", zap.String("topic", topic))

		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[topic] = append(b.subscribers[topic], handler)

	b.logger.Info("Subscribed to topic",
		zap.String("topic", topic),
		zap.Int("total_handlers", len(b.subscribers[topic])))
}

// Publish 发布消息
func (b *ConcurrentMessageBus) Publish(ctx context.Context, topic string, payload interface{}) error {
	if topic == "" {
		b.logger.Warn("主题名称为空")

		return errors.New("主题名称为空")
	}

	select {
	case b.queue <- Message{Topic: topic, Payload: payload, Ctx: ctx, Done: nil}:
		// 消息成功入队
	case <-b.ctx.Done():
		b.logger.Warn("消息总线已关闭，消息被丢弃", zap.String("topic", topic))
	case <-ctx.Done():
		b.logger.Warn("上下文已取消，消息被丢弃", zap.String("topic", topic))
	default:
		b.logger.Warn("队列已满，消息被丢弃", zap.String("topic", topic))

		return errors.New("队列已满")
	}

	return nil
}

// PublishSync 同步发布消息，等待消息处理完成后返回
func (b *ConcurrentMessageBus) PublishSync(ctx context.Context, topic string, payload interface{}) error {
	if topic == "" {
		b.logger.Warn("主题名称为空")
		return errors.New("主题名称为空")
	}

	// 创建完成通知 channel
	done := make(chan struct{})
	msg := Message{Topic: topic, Payload: payload, Ctx: ctx, Done: done}

	select {
	case b.queue <- msg:
		// 消息成功入队，等待处理完成
		select {
		case <-done:
			// 消息处理完成
			return nil
		case <-b.ctx.Done():
			b.logger.Warn("消息总线已关闭，等待被中断", zap.String("topic", topic))
			return errors.New("消息总线已关闭")
		case <-ctx.Done():
			b.logger.Warn("上下文已取消，等待被中断", zap.String("topic", topic))
			return ctx.Err()
		}
	case <-b.ctx.Done():
		b.logger.Warn("消息总线已关闭，消息被丢弃", zap.String("topic", topic))
		return errors.New("消息总线已关闭")
	case <-ctx.Done():
		b.logger.Warn("上下文已取消，消息被丢弃", zap.String("topic", topic))
		return ctx.Err()
	default:
		b.logger.Warn("队列已满，消息被丢弃", zap.String("topic", topic))
		return errors.New("队列已满")
	}
}

// Close 关闭消息总线
func (b *ConcurrentMessageBus) Close() {
	b.logger.Info("正在关闭消息总线...")
	b.cancel()
	b.wg.Wait()
	close(b.queue)
	b.logger.Info("消息总线已关闭")
}

// QueueLength 获取队列长度
func (b *ConcurrentMessageBus) QueueLength() int {
	return len(b.queue)
}

// WorkerCount 获取worker数量
func (b *ConcurrentMessageBus) WorkerCount() int {
	return b.workerCount
}

// GetSubscriberCount 获取指定主题的订阅者数量
func (b *ConcurrentMessageBus) GetSubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers[topic])
}

// GetAllTopics 获取所有已订阅的主题
func (b *ConcurrentMessageBus) GetAllTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.subscribers))
	for topic := range b.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

// Stats 获取统计信息
func (b *ConcurrentMessageBus) Stats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topicStats := make(map[string]int)
	totalHandlers := 0

	for topic, handlers := range b.subscribers {
		count := len(handlers)
		topicStats[topic] = count
		totalHandlers += count
	}

	return map[string]interface{}{
		"worker_count":   b.workerCount,
		"buffer_size":    b.bufferSize,
		"queue_length":   len(b.queue),
		"total_topics":   len(b.subscribers),
		"total_handlers": totalHandlers,
		"topic_handlers": topicStats,
	}
}

// GetRuntimeStats 获取详细的运行时状态统计
func (b *ConcurrentMessageBus) GetRuntimeStats() RuntimeStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 统计主题和处理器
	topicStats := make(map[string]int)
	totalHandlers := 0
	for topic, handlers := range b.subscribers {
		count := len(handlers)
		topicStats[topic] = count
		totalHandlers += count
	}

	// 统计worker状态
	var idleWorkers, processingWorkers, stoppingWorkers, stoppedWorkers int
	workers := make([]WorkerInfo, len(b.workers))

	for i, worker := range b.workers {
		info := worker.getInfo()
		workers[i] = info

		switch info.Status {
		case WorkerStatusIdle:
			idleWorkers++
		case WorkerStatusProcessing:
			processingWorkers++
		case WorkerStatusStopping:
			stoppingWorkers++
		case WorkerStatusStopped:
			stoppedWorkers++
		}
	}

	return RuntimeStats{
		WorkerCount:       b.workerCount,
		BufferSize:        b.bufferSize,
		QueueLength:       len(b.queue),
		IdleWorkers:       idleWorkers,
		ProcessingWorkers: processingWorkers,
		StoppingWorkers:   stoppingWorkers,
		StoppedWorkers:    stoppedWorkers,
		TotalTopics:       len(b.subscribers),
		TotalHandlers:     totalHandlers,
		TopicHandlers:     topicStats,
		Workers:           workers,
	}
}

// GetWorkerInfo 获取指定worker的信息
func (b *ConcurrentMessageBus) GetWorkerInfo(workerID int) (WorkerInfo, bool) {
	if workerID < 1 || workerID > len(b.workers) {
		return WorkerInfo{}, false
	}

	worker := b.workers[workerID-1] // workerID从1开始，数组从0开始
	return worker.getInfo(), true
}

// GetAllWorkerInfo 获取所有worker的信息
func (b *ConcurrentMessageBus) GetAllWorkerInfo() []WorkerInfo {
	workers := make([]WorkerInfo, len(b.workers))
	for i, worker := range b.workers {
		workers[i] = worker.getInfo()
	}
	return workers
}
