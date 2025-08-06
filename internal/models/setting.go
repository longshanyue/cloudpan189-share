package models

import "time"

type Setting struct {
	ID                       int64     `gorm:"primaryKey" json:"id"`
	Title                    string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	EnableAuth               bool      `gorm:"column:enable_auth;type:tinyint(1);default:1" json:"enableAuth"` // 是否启用鉴权 1 启用 0 不启用
	SaltKey                  string    `gorm:"column:salt_key;type:varchar(255);not null" json:"-"`
	LocalProxy               bool      `gorm:"column:local_proxy;type:tinyint(1);default:0" json:"localProxy"`                                // 是否启用本地代理
	MultipleStream           bool      `gorm:"column:multiple_stream;type:tinyint(1);default:0" json:"multipleStream"`                        // 多线程流加速
	BaseURL                  string    `gorm:"column:base_url;type:varchar(255);not null;default:''" json:"baseURL"`                          // base url
	EnableTopFileAutoRefresh bool      `gorm:"column:enable_top_file_auto_refresh;type:tinyint(1);default:1" json:"enableTopFileAutoRefresh"` // 挂载文件自动刷新
	Initialized              bool      `gorm:"column:initialized;type:tinyint(1);default:0" json:"initialized"`                               // 是否初始化完成
	JobThreadCount           int       `gorm:"column:job_thread_count;type:tinyint(1);default:1" json:"jobThreadCount"`                       // 任务线程数
	AutoRefreshMinutes       int       `gorm:"column:auto_refresh_minutes;type:tinyint(1);default:10" json:"autoRefreshMinutes"`              // 自动刷新间隔
	CreatedAt                time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *Setting) TableName() string {
	return "setting"
}
