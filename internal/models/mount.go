package models

import (
	"time"

	"gorm.io/datatypes"
)

type StorageEntry struct {
	ID         int64             `gorm:"primaryKey" json:"id"`
	Name       string            `gorm:"column:name;type:varchar(255);not null" json:"name"`
	LocalPath  string            `gorm:"column:local_path;type:varchar(1024);not null;uniqueIndex:local_path_unique" json:"localPath"`
	Status     string            `gorm:"column:status;type:varchar(20);default:'offline'" json:"status"`       // 状态: online/offline/error
	Protocol   string            `gorm:"column:protocol;type:varchar(20);default:'subscribe'" json:"protocol"` // 来源类型： subscribe 订阅/share 分享
	Addition   datatypes.JSONMap `gorm:"column:addition;type:json" json:"addition"`
	CloudToken int64             `gorm:"column:cloud_token;type:bigint(20);not null;default:0" json:"cloudToken"`
	CreatedAt  time.Time         `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt  time.Time         `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *StorageEntry) TableName() string {
	return "storage_entries"
}

const (
	OsTypeFolder         = "folder"
	OsTypeFile           = "file"
	OsTypeSubscribe      = "subscribe"
	OsTypeSubscribeShare = "subscribe_share"
	OsTypeShare          = "share"
)

type VirtualFile struct {
	ID         int64             `gorm:"primaryKey" json:"id"`
	ParentId   int64             `gorm:"column:parent_id;type:bigint(20);not null;default:0;uniqueIndex:parent_name_unique" json:"parentId"`
	Name       string            `gorm:"column:name;type:varchar(1024);not null;uniqueIndex:parent_name_unique" json:"name"`
	IsTop      int8              `gorm:"column:is_top;type:tinyint(1);default:0" json:"isTop"` // 是否最顶层文件夹
	Size       int64             `gorm:"column:size;type:bigint(20);default:0" json:"size"`
	IsFolder   int8              `gorm:"column:is_folder;type:tinyint(1);default:0" json:"isFolder"`
	Hash       string            `gorm:"column:hash;type:varchar(64);default:''" json:"hash"`
	CreateDate string            `gorm:"column:create_date;type:varchar(20);default:CURRENT_TIMESTAMP" json:"createDate"`
	ModifyDate string            `gorm:"column:modify_date;type:varchar(20);default:CURRENT_TIMESTAMP" json:"modifyDate"`
	OsType     string            `gorm:"column:os_type;type:varchar(20);default:'folder'" json:"osType"` // 读取文件的方式
	Addition   datatypes.JSONMap `gorm:"column:addition;type:json" json:"addition"`
	Rev        string            `gorm:"column:rev;type:varchar(64);default:''" json:"rev"` // 版本 用于下次扫描时知道当前文件是删除还是修改还是新增
	//IsDelete   int8              `gorm:"column:is_delete;type:tinyint(1);default:0" json:"-"` // 删除标记 延迟删除
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *VirtualFile) TableName() string {
	return "virtual_files"
}
