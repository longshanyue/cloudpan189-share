package models

import (
	"encoding/json"
	"strconv"
	"time"

	"gorm.io/gorm"
)

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

type SettingDictValue string

func (s SettingDictValue) Value() string {
	return string(s)
}

func (s SettingDictValue) Int() int {
	i, _ := strconv.Atoi(string(s))

	return i
}

func (s SettingDictValue) Int64() int64 {
	i, _ := strconv.ParseInt(string(s), 10, 64)

	return i
}

func (s SettingDictValue) Bool() bool {
	b, _ := strconv.ParseBool(string(s))

	return b
}

func (s SettingDictValue) Float64() float64 {
	f, _ := strconv.ParseFloat(string(s), 64)

	return f
}

func (s SettingDictValue) StringSlice() []string {
	var slice []string

	_ = json.Unmarshal([]byte(string(s)), &slice)

	return slice
}

type SettingDict struct {
	ID        int64            `gorm:"primaryKey" json:"id"`
	Key       string           `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value     SettingDictValue `gorm:"column:value;type:varchar(255);not null" json:"value"`
	Type      string           `gorm:"column:type;type:varchar(255);not null" json:"type"`
	CreatedAt time.Time        `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time        `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (s *SettingDict) TableName() string {
	return "setting_dicts"
}

const (
	SettingDictKeyMultipleStreamThreadCount = "multiple_stream_thread_count"
	SettingDictKeyMultipleStreamChunkSize   = "multiple_stream_chunk_size"
	SettingDictKeyStrmFileEnable            = "strm_file_enable"
	SettingDictKeyStrmSupportFileExtList    = "strm_support_file_ext_list"
)

const (
	DefaultMultipleStreamThreadCount = 6
	DefaultMultipleStreamChunkSize   = 1024 * 1024 * 4
	DefaultStrmFileEnable            = false
)

var (
	DefaultStrmSupportFileExtList = []string{
		"mp4", "mkv", "avi", "mov", "wmv", "flv", "webm", "m4v",
		"mpg", "mpeg", "m2v", "m4p", "m4b", "ts", "mts", "m2ts", "m2t",
		"mxf", "dv", "dvr-ms", "asf", "3gp", "3g2", "f4v", "f4p", "f4a", "f4b",
		"vob", "ogv", "ogg", "divx", "xvid", "rm", "rmvb", "dat", "nsv",
		"qt", "amv", "mpv", "m1v", "svi", "viv", "fli", "flc",
	}
)

func (s *SettingDict) query(db *gorm.DB, key string) (string, error) {
	var m = new(SettingDict)

	if err := db.Where("key", key).First(m).Error; err != nil {
		return "", err
	}

	return string(m.Value), nil
}

func (s *SettingDict) store(db *gorm.DB, key, value string, typ string) *gorm.DB {
	var count int64

	if result := db.Model(s).Where("key", key).Count(&count); result.Error != nil {
		return result
	}

	if count > 0 {
		return db.Model(s).Where("key", key).Update("value", value)
	}

	return db.Create(&SettingDict{
		Key:   key,
		Value: SettingDictValue(value),
		Type:  typ,
	})
}

func (s *SettingDict) GetMultipleStreamThreadCount(db *gorm.DB) int {
	value, err := s.query(db, SettingDictKeyMultipleStreamThreadCount)
	if err != nil {
		return DefaultMultipleStreamThreadCount
	}

	var v int64

	if v, err = strconv.ParseInt(value, 10, 64); err != nil {
		return DefaultMultipleStreamThreadCount
	}

	return int(v)
}

func (s *SettingDict) SetMultipleStreamThreadCount(db *gorm.DB, value int) *gorm.DB {
	return s.store(db, SettingDictKeyMultipleStreamThreadCount, strconv.FormatInt(int64(value), 10), "int")
}

func (s *SettingDict) GetMultipleStreamChunkSize(db *gorm.DB) int64 {
	value, err := s.query(db, SettingDictKeyMultipleStreamChunkSize)
	if err != nil {
		return DefaultMultipleStreamChunkSize
	}

	var v int64

	if v, err = strconv.ParseInt(value, 10, 64); err != nil {
		return DefaultMultipleStreamChunkSize
	}

	return v
}

func (s *SettingDict) SetMultipleStreamChunkSize(db *gorm.DB, value int64) *gorm.DB {
	return s.store(db, SettingDictKeyMultipleStreamChunkSize, strconv.FormatInt(value, 10), "int64")
}

func (s *SettingDict) GetStrmFileEnable(db *gorm.DB) bool {
	value, err := s.query(db, SettingDictKeyStrmFileEnable)
	if err != nil {
		return DefaultStrmFileEnable
	}

	var v bool

	if v, err = strconv.ParseBool(value); err != nil {
		return DefaultStrmFileEnable
	}

	return v
}

func (s *SettingDict) SetStrmFileEnable(db *gorm.DB, value bool) *gorm.DB {
	return s.store(db, SettingDictKeyStrmFileEnable, strconv.FormatBool(value), "bool")
}

func (s *SettingDict) GetStrmSupportFileExtList(db *gorm.DB) []string {
	value, err := s.query(db, SettingDictKeyStrmSupportFileExtList)
	if err != nil {
		return DefaultStrmSupportFileExtList
	}

	var v []string

	if err = json.Unmarshal([]byte(value), &v); err != nil {
		return DefaultStrmSupportFileExtList
	}

	return v
}

func (s *SettingDict) SetStrmSupportFileExtList(db *gorm.DB, value []string) *gorm.DB {
	b, _ := json.Marshal(value)

	return s.store(db, SettingDictKeyStrmSupportFileExtList, string(b), "json")
}
