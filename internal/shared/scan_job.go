package shared

import (
	"sync"

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
	ErrScanJobFull = errors.New("扫描任务队列已满，请稍后再试")
)

type ScanJobType string

const (
	ScanJobTypeDel         ScanJobType = "del"
	ScanJobTypeRefresh     ScanJobType = "refresh"      // 添加时或普通刷新时调用
	ScanJobTypeDeepRefresh ScanJobType = "deep_refresh" // 递归刷新时调用 rev 相同也继续扫描
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
	return scanJobInstance.Publish(typ, msg)
}

func ScanJobRead() <-chan ScanMsg {
	return scanJobInstance.Read()
}
