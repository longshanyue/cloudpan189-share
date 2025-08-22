package eventbus

import "context"

// Handler 事件处理函数
type Handler func(ctx context.Context, data interface{}) error

// Subscription 订阅接口
type Subscription interface {
	ID() string
	Topic() string
	Close()
}

// EventBus 事件总线接口
type EventBus interface {
	Subscribe(topic string, handler Handler) Subscription
	Unsubscribe(sub Subscription)
	Publish(ctx context.Context, topic string, data interface{}) error
	PublishSync(ctx context.Context, topic string, data interface{}) error
	Close()
}
