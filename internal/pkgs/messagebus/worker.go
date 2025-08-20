package messagebus

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type worker struct {
	id              int
	bus             *ConcurrentMessageBus
	wg              *sync.WaitGroup
	status          int32 // 使用atomic操作的状态
	currentTopic    string
	processingStart *time.Time
	totalProcessed  int64
	lastProcessedAt *time.Time
	mu              sync.RWMutex // 保护非原子字段
}

func newWorker(id int, bus *ConcurrentMessageBus, wg *sync.WaitGroup) *worker {
	w := &worker{
		id:  id,
		bus: bus,
		wg:  wg,
	}
	// 初始状态为空闲
	atomic.StoreInt32(&w.status, int32(WorkerStatusIdle))
	return w
}

// setStatus 设置worker状态
func (w *worker) setStatus(status WorkerStatus) {
	atomic.StoreInt32(&w.status, int32(status))
}

// getStatus 获取worker状态
func (w *worker) getStatus() WorkerStatus {
	return WorkerStatus(atomic.LoadInt32(&w.status))
}

// getInfo 获取worker信息
func (w *worker) getInfo() WorkerInfo {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	info := WorkerInfo{
		ID:              w.id,
		Status:          w.getStatus(),
		CurrentTopic:    w.currentTopic,
		TotalProcessed:  atomic.LoadInt64(&w.totalProcessed),
	}
	
	if w.processingStart != nil {
		info.ProcessingStart = w.processingStart
	}
	if w.lastProcessedAt != nil {
		info.LastProcessedAt = w.lastProcessedAt
	}
	
	return info
}

// setCurrentTopic 设置当前处理的主题
func (w *worker) setCurrentTopic(topic string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.currentTopic = topic
}

// setProcessingStart 设置开始处理时间
func (w *worker) setProcessingStart(t *time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.processingStart = t
}

// incrementProcessed 增加处理计数
func (w *worker) incrementProcessed() {
	atomic.AddInt64(&w.totalProcessed, 1)
	now := time.Now()
	w.mu.Lock()
	w.lastProcessedAt = &now
	w.mu.Unlock()
}

func (w *worker) start() {
	w.wg.Add(1)
	go w.run()
}

func (w *worker) run() {
	defer func() {
		w.setStatus(WorkerStatusStopped)
		w.wg.Done()
	}()

	w.bus.logger.Info("工作线程已启动", zap.Int("worker_id", w.id))
	w.setStatus(WorkerStatusIdle)

	for {
		select {
		case msg := <-w.bus.queue:
			w.handleMessage(msg)
			w.setStatus(WorkerStatusIdle) // 处理完成后回到空闲状态
		case <-w.bus.ctx.Done():
			w.setStatus(WorkerStatusStopping)
			w.bus.logger.Info("工作线程正在停止...", zap.Int("worker_id", w.id))
			// 处理剩余消息
			for {
				select {
				case msg := <-w.bus.queue:
					w.handleMessage(msg)
				default:
					w.bus.logger.Info("工作线程已停止", zap.Int("worker_id", w.id))
					return
				}
			}
		}
	}
}

func (w *worker) handleMessage(msg Message) {
	// 设置处理状态
	w.setStatus(WorkerStatusProcessing)
	w.setCurrentTopic(msg.Topic)
	now := time.Now()
	w.setProcessingStart(&now)
	
	// 确保在函数结束时发送完成通知（如果需要）和清理状态
	defer func() {
		w.incrementProcessed()
		w.setCurrentTopic("")
		w.setProcessingStart(nil)
		
		if msg.Done != nil {
			close(msg.Done)
		}
	}()

	w.bus.mu.RLock()
	handlers := make([]Handler, len(w.bus.subscribers[msg.Topic]))
	copy(handlers, w.bus.subscribers[msg.Topic])
	w.bus.mu.RUnlock()

	if len(handlers) == 0 {
		w.bus.logger.Debug("主题没有处理器",
			zap.Int("worker_id", w.id),
			zap.String("topic", msg.Topic))
		return
	}

	// 顺序执行所有处理器
	for i, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					w.bus.logger.Error("处理器发生异常",
						zap.Int("worker_id", w.id),
						zap.Int("handler_index", i),
						zap.String("topic", msg.Topic),
						zap.Any("panic", r))
				}
			}()

			handler(msg.Ctx, msg.Payload)
		}()
	}
}
