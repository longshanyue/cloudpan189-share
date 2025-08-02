package models

import "time"

const (
	PermissionBase = 1 << iota
	PermissionDavRead
	PermissionAdmin
)

type User struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"column:username;type:varchar(255);not null;uniqueIndex" json:"username"`
	Password    string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Status      int8      `gorm:"column:status;type:tinyint(1);default:1" json:"status"`
	Permissions uint8     `gorm:"column:permissions;type:tinyint(1);default:1" json:"permissions"`
	Version     int       `gorm:"column:version;type:int(11);default:1" json:"version"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (u *User) TableName() string {
	return "users"
}
