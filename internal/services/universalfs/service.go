package universalfs

import (
	"github.com/xxcheng123/cloudpan189-share/internal/fs"
	"golang.org/x/net/webdav"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type Service interface {
	Open(prefix string, format string) gin.HandlerFunc
	FileDownload() gin.HandlerFunc
	DavMiddleware() gin.HandlerFunc
	BaseMiddleware() gin.HandlerFunc
	Put() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type service struct {
	db         *gorm.DB
	logger     *zap.Logger
	startTime  time.Time
	cache      *cache.Cache
	g          singleflight.Group
	LockSystem webdav.LockSystem
	fs         fs.FS
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:         db,
		logger:     logger,
		startTime:  time.Now(),
		cache:      cache.New(time.Minute, time.Minute*10),
		LockSystem: webdav.NewMemLS(),
		fs:         fs.NewFS(db, logger),
	}
}

type FileInfo struct {
	*models.VirtualFile
	Path        string      `json:"path"`
	Href        string      `json:"href"`
	Children    []*FileInfo `json:"children,omitempty"`
	DownloadURL string      `json:"downloadURL,omitempty"`
}

type ReadSession struct {
	CloudTokenID int64 `json:"cloudTokenId"`
}
