package models

import "time"

type CloudToken struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	AccessToken string    `gorm:"column:access_token;type:varchar(255);not null" json:"accessToken"`
	ExpiresIn   int64     `gorm:"column:expires_in;type:bigint(20);not null" json:"expiresIn"`
	Status      int8      `gorm:"column:status;type:tinyint(1);default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (c *CloudToken) TableName() string {
	return "cloud_tokens"
}
