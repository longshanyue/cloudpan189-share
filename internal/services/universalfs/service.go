package universalfs

import (
	"crypto/md5"
	"encoding/hex"
	"golang.org/x/net/webdav"
	"net/url"
	"sort"
	"strconv"
	"strings"
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
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:         db,
		logger:     logger,
		startTime:  time.Now(),
		cache:      cache.New(time.Minute, time.Minute*10),
		LockSystem: webdav.NewMemLS(),
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

func enc(values url.Values, key string) url.Values {
	if !values.Has("timestamp") {
		timestamp := time.Now().Add(time.Hour * 6).Unix()
		values.Set("timestamp", strconv.FormatInt(timestamp, 10))
	}

	// 排序并生成签名字符串
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// 构建待签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k + "=" + values.Get(k) + "&")
	}
	signStr.WriteString("key=" + key)

	// 计算MD5
	h := md5.New()
	h.Write([]byte(signStr.String()))
	sign := hex.EncodeToString(h.Sum(nil))

	values.Set("sign", sign)
	return values
}

// 验证签名是否有效
func verify(values url.Values, key string) bool {
	// 获取并移除签名
	sign := values.Get("sign")
	values.Del("sign")

	// 重新计算签名
	newValues := enc(values, key)
	newSign := newValues.Get("sign")

	// 比较签名是否一致
	return sign == newSign
}
