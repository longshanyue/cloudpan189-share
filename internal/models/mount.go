package models

import (
	"time"

	"gorm.io/datatypes"
)

type OsType = string

const (
	OsTypeFolder         = "folder"
	OsTypeFile           = "file"
	OsTypeSubscribe      = "subscribe"
	OsTypeSubscribeShare = "subscribe_share"
	OsTypeShare          = "share"
	OsTypeRealFile       = "real_file"
	OsTypeStrmFile       = "strm_file"
)

func OsTypeAllowDelete(osType OsType) bool {
	return osType == OsTypeRealFile || osType == OsTypeStrmFile
}

type VirtualFile struct {
	ID         int64             `gorm:"primaryKey" json:"id"`
	ParentId   int64             `gorm:"column:parent_id;type:bigint(20);not null;default:0;uniqueIndex:parent_name_unique" json:"parentId"`
	LinkId     int64             `gorm:"column:link_id;type:bigint(20);default:0;index:link_id_index" json:"linkId"` // 关联id，用于strm文件，当文件被删除后，实现关联的 strm 文件快速删除
	Name       string            `gorm:"column:name;type:varchar(1024);not null;uniqueIndex:parent_name_unique" json:"name"`
	IsTop      int8              `gorm:"column:is_top;type:tinyint(1);default:0" json:"isTop"` // 是否最顶层文件夹
	Size       int64             `gorm:"column:size;type:bigint(20);default:0" json:"size"`
	IsFolder   int8              `gorm:"column:is_folder;type:tinyint(1);default:0" json:"isFolder"`
	Hash       string            `gorm:"column:hash;type:varchar(64);default:''" json:"hash"`
	CreateDate string            `gorm:"column:create_date;type:varchar(20);default:CURRENT_TIMESTAMP" json:"createDate"`
	ModifyDate string            `gorm:"column:modify_date;type:varchar(20);default:CURRENT_TIMESTAMP" json:"modifyDate"`
	OsType     OsType            `gorm:"column:os_type;type:varchar(20);default:'folder'" json:"osType"` // 读取文件的方式
	Addition   datatypes.JSONMap `gorm:"column:addition;type:json" json:"addition"`
	Rev        string            `gorm:"column:rev;type:varchar(64);default:''" json:"rev"` // 版本 用于下次扫描时知道当前文件是删除还是修改还是新增
	//IsDelete   int8              `gorm:"column:is_delete;type:tinyint(1);default:0" json:"-"` // 删除标记 延迟删除
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *VirtualFile) TableName() string {
	return "virtual_files"
}
