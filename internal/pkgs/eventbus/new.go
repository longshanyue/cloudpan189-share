package eventbus

// New 创建事件总线
func New() EventBus {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig 使用配置创建事件总线
func NewWithConfig(config *Config) EventBus {
	if config == nil {
		config = DefaultConfig()
	}

	eb := &eventBus{
		subscribers: make(map[string][]*subscription),
		config:      config,
		eventCh:     make(chan *Event, config.BufferSize),
		done:        make(chan struct{}),
		concurrency: make(chan struct{}, config.MaxConcurrency),
	}

	// 启动全局事件处理器
	go eb.processor()

	return eb
}
