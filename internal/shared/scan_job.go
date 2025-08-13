package shared

import (
	"context"
	"fmt"
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
	ErrScanJobFull   = errors.New("执行任务队列已满，请稍后再试")
	ErrJobConflict   = errors.New("当前任务或与已存在任务冲突")
	ErrJobNotFound   = errors.New("任务未找到")
	ErrInvalidStatus = fmt.Errorf("invalid job status transition")
)

type ScanJobType string

const (
	ScanJobTypeDel         ScanJobType = "del"
	ScanJobTypeRefresh     ScanJobType = "refresh"      // 添加时或普通刷新时调用
	ScanJobTypeDeepRefresh ScanJobType = "deep_refresh" // 递归刷新时调用 rev 相同也继续扫描
	ScanJobRebuildStrm     ScanJobType = "rebuild_strm"
	ScanJobClearStrm       ScanJobType = "clear_strm"
	ScanJobClearRealFile   ScanJobType = "clear_real_file"
	ScanJobTypeScanTop     ScanJobType = "scan_top" // 扫描所有顶层文件
)

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

// RunningStat 将任务状态设置为运行中
func RunningStat(fileId int64) error {
	v, ok := jobMap.Load(fileId)
	if !ok {
		if err := InitStat(fileId, ScanJobTypeRefresh); err != nil {
			return err
		}
		return RunningStat(fileId)
	}

	jobStat := v.(*JobStat)

	// 检查状态转换是否合法
	if jobStat.Status != JobStatusWaiting {
		return fmt.Errorf("%w: job is already running", ErrInvalidStatus)
	}

	jobStat.Status = JobStatusRunning
	jobStat.StartTime = time.Now()

	return nil
}

// FinishStat 完成任务并删除统计信息
func FinishStat(fileId int64) {
	v, ok := jobMap.Load(fileId)
	if !ok {
		return // 任务不存在，直接返回
	}

	jobStat := v.(*JobStat)

	// 取消context
	if jobStat.cancel != nil {
		jobStat.cancel()
	}

	// 删除任务
	jobMap.Delete(fileId)
}

// CancelStat 取消任务
func CancelStat(fileId int64) error {
	v, ok := jobMap.Load(fileId)
	if !ok {
		return ErrJobNotFound
	}

	jobStat := v.(*JobStat)

	// 取消context
	if jobStat.cancel != nil {
		jobStat.cancel()
	}

	// 删除任务
	jobMap.Delete(fileId)
	return nil
}

// GetStat 获取指定任务的统计信息
func GetStat(fileId int64) (*JobStat, error) {
	v, ok := jobMap.Load(fileId)
	if !ok {
		return nil, ErrJobNotFound
	}

	jobStat := v.(*JobStat)

	// 返回副本，避免外部修改
	result := &JobStat{
		FileID:       jobStat.FileID,
		JobType:      jobStat.JobType,
		Status:       jobStat.Status,
		GenerateTime: jobStat.GenerateTime,
		StartTime:    jobStat.StartTime,
		WaitCount:    jobStat.WaitCount,
		ScannedCount: jobStat.ScannedCount,
		// 不复制私有字段
	}

	return result, nil
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

// GetJobContext 获取任务的context（用于取消操作）
func GetJobContext(fileId int64) (context.Context, error) {
	v, ok := jobMap.Load(fileId)
	if !ok {
		return nil, ErrJobNotFound
	}

	jobStat := v.(*JobStat)
	return jobStat.GetContext(), nil
}

// UpdateJobProgress 更新任务进度的便捷方法
func UpdateJobProgress(fileId int64, scannedCount, waitCount int64) error {
	v, ok := jobMap.Load(fileId)
	if !ok {
		return ErrJobNotFound
	}

	jobStat := v.(*JobStat)
	jobStat.UpdateProgress(scannedCount, waitCount)
	return nil
}

// ScanTopJobPublish 发布扫描顶层文件任务的便捷方法
func ScanTopJobPublish() error {
	// 使用一个虚拟的文件对象，ID为0表示扫描所有顶层文件
	virtualFile := &models.VirtualFile{ID: 0}

	return ScanJobPublish(ScanJobTypeScanTop, virtualFile)
}
