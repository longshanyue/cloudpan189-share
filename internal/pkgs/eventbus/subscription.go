package eventbus

import (
	"sync/atomic"
)

// subscription 订阅实现
type subscription struct {
	id      string
	topic   string
	handler Handler
	closed  int32
}

// newSubscription 创建订阅
func newSubscription(id, topic string, handler Handler) *subscription {
	return &subscription{
		id:      id,
		topic:   topic,
		handler: handler,
	}
}

func (s *subscription) ID() string {
	return s.id
}

func (s *subscription) Topic() string {
	return s.topic
}

func (s *subscription) Close() {
	atomic.StoreInt32(&s.closed, 1)
}

func (s *subscription) isClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}
