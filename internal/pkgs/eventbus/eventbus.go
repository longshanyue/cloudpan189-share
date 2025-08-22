package eventbus

import (
	"context"
	"sync"
	"sync/atomic"
)

// eventBus 事件总线实现
type eventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	counter     int64
	closed      int32
	config      *Config

	// 全局事件处理
	eventCh     chan *Event
	done        chan struct{}
	concurrency chan struct{}  // 全局并发控制信号量
	wg          sync.WaitGroup // 等待所有处理完成
}

// Subscribe 订阅
func (eb *eventBus) Subscribe(topic string, handler Handler) Subscription {
	if atomic.LoadInt32(&eb.closed) == 1 {
		return nil
	}

	id := generateID(atomic.AddInt64(&eb.counter, 1))
	sub := newSubscription(id, topic, handler)

	eb.mu.Lock()
	eb.subscribers[topic] = append(eb.subscribers[topic], sub)
	eb.mu.Unlock()

	return sub
}

// Unsubscribe 取消订阅
func (eb *eventBus) Unsubscribe(sub Subscription) {
	if sub == nil {
		return
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	topic := sub.Topic()
	subs := eb.subscribers[topic]

	for i, s := range subs {
		if s.ID() == sub.ID() {
			s.Close()
			eb.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	if len(eb.subscribers[topic]) == 0 {
		delete(eb.subscribers, topic)
	}
}

// Publish 异步发布，返回错误
func (eb *eventBus) Publish(ctx context.Context, topic string, data interface{}) error {
	if atomic.LoadInt32(&eb.closed) == 1 {
		return ErrEventBusClosed
	}

	// 检查context是否已经取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	eb.mu.RLock()
	subs := make([]*subscription, len(eb.subscribers[topic]))
	copy(subs, eb.subscribers[topic])
	eb.mu.RUnlock()

	if len(subs) == 0 {
		return ErrNoSubscribers
	}

	var successCount, failedCount int
	var lastError error

	// 为每个订阅者创建事件
	for _, sub := range subs {
		if sub.isClosed() {
			failedCount++
			lastError = ErrSubscriptionClosed
			continue
		}

		event := &Event{
			Topic:   topic,
			Data:    data,
			Context: ctx,
			Handler: sub.handler,
		}

		// 发送到全局事件队列
		select {
		case eb.eventCh <- event:
			successCount++
		case <-ctx.Done():
			failedCount++
			lastError = ctx.Err()
		default:
			// 队列满，返回错误
			failedCount++
			lastError = ErrChannelFull
		}
	}

	// 如果有失败的，返回错误
	if failedCount > 0 {
		return &PublishError{
			Topic:        topic,
			SuccessCount: successCount,
			FailedCount:  failedCount,
			LastError:    lastError,
		}
	}

	return nil
}

// PublishSync 同步发布
func (eb *eventBus) PublishSync(ctx context.Context, topic string, data interface{}) error {
	if atomic.LoadInt32(&eb.closed) == 1 {
		return ErrEventBusClosed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	eb.mu.RLock()
	subs := make([]*subscription, len(eb.subscribers[topic]))
	copy(subs, eb.subscribers[topic])
	eb.mu.RUnlock()

	if len(subs) == 0 {
		return ErrNoSubscribers
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(subs))

	for _, sub := range subs {
		if sub.isClosed() {
			errChan <- ErrSubscriptionClosed
			continue
		}

		wg.Add(1)

		event := &Event{
			Topic:   topic,
			Data:    data,
			Context: ctx,
			Handler: sub.handler,
			Result:  make(chan error, 1),
		}

		// 发送到全局事件队列
		select {
		case eb.eventCh <- event:
			// 成功发送，等待结果
			go func() {
				defer wg.Done()
				select {
				case err := <-event.Result:
					if err != nil {
						errChan <- err
					}
				case <-ctx.Done():
					errChan <- ctx.Err()
				}
			}()
		case <-ctx.Done():
			wg.Done()
			errChan <- ctx.Err()
		default:
			// 队列满
			wg.Done()
			errChan <- ErrChannelFull
		}
	}

	// 等待所有完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	close(errChan)

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return &MultiError{Errors: errors}
	}

	return nil
}

// processor 全局事件处理器，控制整个bus的并发
func (eb *eventBus) processor() {
	defer func() {
		// 处理剩余事件
		for {
			select {
			case event := <-eb.eventCh:
				eb.handleEventDirect(event)
			default:
				return
			}
		}
	}()

	for {
		select {
		case event := <-eb.eventCh:
			// 尝试获取并发许可
			select {
			case eb.concurrency <- struct{}{}:
				// 获得许可，启动goroutine处理
				eb.wg.Add(1)
				go eb.handleEventConcurrent(event)
			case <-eb.done:
				return
			default:
				// 无法获得并发许可，直接处理（阻塞当前goroutine）
				eb.handleEventDirect(event)
			}
		case <-eb.done:
			return
		}
	}
}

// handleEventConcurrent 并发处理事件
func (eb *eventBus) handleEventConcurrent(event *Event) {
	defer func() {
		<-eb.concurrency // 释放并发许可
		eb.wg.Done()
	}()

	eb.handleEventDirect(event)
}

// handleEventDirect 直接处理事件
func (eb *eventBus) handleEventDirect(event *Event) {
	var err error

	// 检查context
	select {
	case <-event.Context.Done():
		err = event.Context.Err()
	default:
		// 执行处理器
		err = event.Handler(event.Context, event.Data)
	}

	// 返回结果（同步发布时）
	if event.Result != nil {
		select {
		case event.Result <- err:
		default:
		}
	}
}

// Close 关闭事件总线
func (eb *eventBus) Close() {
	if !atomic.CompareAndSwapInt32(&eb.closed, 0, 1) {
		return
	}

	close(eb.done)
	eb.wg.Wait() // 等待所有处理完成

	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, subs := range eb.subscribers {
		for _, sub := range subs {
			sub.Close()
		}
	}

	eb.subscribers = make(map[string][]*subscription)
}
