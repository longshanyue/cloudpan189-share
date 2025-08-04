package configs

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	logger2 "github.com/xxcheng123/cloudpan189-share/internal/pkgs/logger"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/zeromicro/go-zero/core/conf"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	BuildDate  string
	Commit     string
	GitBranch  string
	GitSummary string
)

var c = new(Config)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "etc/config.yaml", "config path")
}

func init() {
	flag.Parse()

	conf.MustLoad(configPath, c)
}

func Init() {
	var err error

	// 先判断数据库目录是否存在
	dbDir := filepath.Dir(c.DBFile)
	if dbDir != "" && dbDir != "." {
		// 检查目录是否存在
		if _, err = os.Stat(dbDir); os.IsNotExist(err) {
			// 目录不存在，创建目录
			if err = os.MkdirAll(dbDir, 0755); err != nil {
				panic(fmt.Sprintf("创建数据库目录失败: %v", err))
			}
		} else if err != nil {
			// 其他错误（如权限问题）
			panic(fmt.Sprintf("检查数据库目录失败: %v", err))
		}
	}

	db, err = gorm.Open(sqlite.Open(c.DBFile), &gorm.Config{
		NowFunc: func() time.Time {
			// 使用中国时区，处理加载失败的情况
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				// 如果加载失败，使用固定偏移量 UTC+8
				loc = time.FixedZone("CST", 8*3600)
			}

			return time.Now().In(loc)
		},
	})
	if err != nil {
		panic(err)
	}

	// 数据迁移
	if err = db.AutoMigrate(
		new(models.User),
		new(models.Setting),
		new(models.CloudToken),
		new(models.VirtualFile),
	); err != nil {
		panic(err)
	}

	{
		var options = []logger2.Option{
			logger2.WithTimeLayout(time.DateTime),
			logger2.WithFileRotationP(c.LogFile),
			logger2.WithInfoLevel(),
			//logger2.WithDebugLevel(),
			logger2.WithOutputInConsole(),
			logger2.WithField("build_info", fmt.Sprintf("[buildDate:%s]&&[commit:%s]&&[gitSummary:%s]&&[gitBranch:%s]", BuildDate, Commit, GitSummary, GitBranch)),
		}

		logger, err = logger2.NewJSONLogger(options...)
		if err != nil {
			panic(err)
		}
	}

	//initUser()
	initSetting()

	var setting = new(models.Setting)
	if err = db.First(setting).Error; err != nil {
		panic(err)
	}

	shared.Setting = setting

	logger.Info("binary build info",
		zap.String("build date", BuildDate),
		zap.String("go version", runtime.Version()),
		zap.String("git commit", Commit),
		zap.String("git branch", GitBranch),
		zap.String("git summar", GitSummary),
	)
}

func initUser() {
	var count int64

	db.Model(new(models.User)).Count(&count)

	if count > 0 {
		return
	}

	// 随机生成密码
	pass := utils.GenerateRandomPassword(12)
	// md5计算

	user := &models.User{
		Username:    "admin",
		Password:    hash(pass),
		Permissions: models.PermissionAdmin | models.PermissionDavRead | models.PermissionBase,
	}

	logger.Info("init create admin user", zap.String("username", user.Username), zap.String("password", pass))

	db.Create(user)
}

// 生成 md5
func hash(input string) string {
	data := []byte(input)
	has := md5.Sum(data)

	return fmt.Sprintf("%x", has)
}

func initSetting() {
	var count int64

	db.Model(new(models.Setting)).Count(&count)

	if count > 0 {
		return
	}

	user := &models.Setting{
		Title:      "天翼订阅小站",
		EnableAuth: true,
		SaltKey:    utils.GenerateRandomPassword(16),
	}

	db.Create(user)
}
