package eventbus

import "context"

// Event 事件
type Event struct {
	Topic   string
	Data    interface{}
	Context context.Context
	Handler Handler
	Result  chan error // 同步发布时使用
}
