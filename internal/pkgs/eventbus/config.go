package eventbus

// Config 事件总线配置
type Config struct {
	BufferSize     int // channel缓冲区大小
	MaxConcurrency int // 整个bus的最大并发处理数
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		BufferSize:     1000,
		MaxConcurrency: 10, // 整个bus最多10个并发
	}
}
