package models

import "time"

type SystemLog struct {
	ID    int64  `gorm:"primaryKey" json:"id"`
	Type  string `gorm:"column:type;type:varchar(64);default:''" json:"type"`
	Title string `gorm:"column:title;type:varchar(1024);default:''" json:"title"`
	// Cost  ms
	Cost      int64     `gorm:"column:cost;type:bigint(20);default:0" json:"cost"`
	Result    string    `gorm:"column:result;type:varchar(64);default:''" json:"result"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *SystemLog) TableName() string {
	return "system_logs"
}
