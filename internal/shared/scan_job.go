package shared

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
)

type scanJob struct {
	ch chan ScanMsg
	mu sync.Mutex
}

var scanJobInstance = &scanJob{
	ch: make(chan ScanMsg, 8),
}

var (
	ErrScanJobFull = errors.New("执行任务队列已满，请稍后再试")
	ErrJobConflict = errors.New("当前任务或与已存在任务冲突")
)

type ScanJobType string

type ScanMsg struct {
	Type ScanJobType
	Msg  *models.VirtualFile
}

func (s *scanJob) Publish(typ ScanJobType, msg *models.VirtualFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.ch) >= cap(s.ch) {
		return ErrScanJobFull
	}

	s.ch <- ScanMsg{
		Type: typ,
		Msg:  msg,
	}

	return nil
}

func (s *scanJob) Read() <-chan ScanMsg {
	return s.ch
}

func ScanJobPublish(typ ScanJobType, msg *models.VirtualFile) error {
	if err := InitStat(msg.ID, typ); err != nil {
		return err
	}

	return scanJobInstance.Publish(typ, msg)
}

func ScanJobRead() <-chan ScanMsg {
	return scanJobInstance.Read()
}

type JobStatus int8

const (
	JobStatusWaiting JobStatus = iota
	JobStatusRunning
)

var (
	jobMap sync.Map
)

type JobStat struct {
	FileID       int64       `json:"fileId"`
	JobType      ScanJobType `json:"jobType"`
	Status       JobStatus   `json:"status"`
	GenerateTime time.Time   `json:"generateTime"`
	StartTime    time.Time   `json:"startTime"`
	WaitCount    int64       `json:"waitCount,omitempty"`
	ScannedCount int64       `json:"scannedCount,omitempty"`

	// 私有字段
	ctx    context.Context
	cancel context.CancelFunc
}

// UpdateProgress 更新进度
func (js *JobStat) UpdateProgress(scannedCount, waitCount int64) {
	js.ScannedCount += scannedCount
	js.WaitCount += waitCount
}

// GetContext 获取任务的context
func (js *JobStat) GetContext() context.Context {
	return js.ctx
}

// InitStat 初始化任务统计
func InitStat(fileId int64, jobType ScanJobType) error {
	if _, ok := jobMap.Load(fileId); ok {
		return ErrJobConflict
	}

	ctx, cancel := context.WithCancel(context.Background())

	jobStat := &JobStat{
		FileID:       fileId,
		JobType:      jobType,
		Status:       JobStatusWaiting,
		GenerateTime: time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}

	jobMap.Store(fileId, jobStat)
	return nil
}

// ListStat 获取所有任务统计信息
func ListStat() []*JobStat {
	var list []*JobStat
	jobMap.Range(func(key, value interface{}) bool {
		jobStat := value.(*JobStat)

		// 创建副本
		statCopy := &JobStat{
			FileID:       jobStat.FileID,
			JobType:      jobStat.JobType,
			Status:       jobStat.Status,
			GenerateTime: jobStat.GenerateTime,
			StartTime:    jobStat.StartTime,
			WaitCount:    jobStat.WaitCount,
			ScannedCount: jobStat.ScannedCount,
			// 不复制私有字段
		}

		list = append(list, statCopy)
		return true
	})

	return list
}
