package shared

import "github.com/xxcheng123/cloudpan189-share/internal/models"

var Setting = &models.Setting{}

var (
	MultipleStreamThreadCount int   = models.DefaultMultipleStreamThreadCount
	MultipleStreamChunkSize   int64 = models.DefaultMultipleStreamChunkSize
)
