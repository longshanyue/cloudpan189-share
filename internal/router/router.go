package router

import (
	"fmt"
	"github.com/xxcheng123/cloudpan189-share/internal/services/usergroup"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	embed "github.com/xxcheng123/cloudpan189-share"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	settingS "github.com/xxcheng123/cloudpan189-share/internal/services/setting"
	"github.com/xxcheng123/cloudpan189-share/internal/services/storage"
	"github.com/xxcheng123/cloudpan189-share/internal/services/universalfs"
	"github.com/xxcheng123/cloudpan189-share/internal/services/user"
	"go.uber.org/zap"
)

func StartHTTPServer() error {
	var (
		engine = gin.Default()
		db     = configs.DB()
		logger = configs.Logger()
		config = configs.GetConfig()
	)

	var (
		userService        = user.NewService(db, logger)
		cloudTokenService  = cloudtoken.NewService(db, logger)
		settingService     = settingS.NewService(db, logger)
		storageService     = storage.NewService(db, logger)
		universalFsService = universalfs.NewService(db, logger)
		userGroupService   = usergroup.NewService(db, logger)
	)

	openapiRouter := engine.Group("/api")
	userRouter := openapiRouter.Group("/user")
	{
		userRouter.POST("/login", userService.Login())
		userRouter.POST("/refresh_token", userService.RefreshToken())

		userRouter.POST("/add", userService.AuthMiddleware(models.PermissionAdmin), userService.Add())
		userRouter.POST("/del", userService.AuthMiddleware(models.PermissionAdmin), userService.Del())
		userRouter.POST("/update", userService.AuthMiddleware(models.PermissionAdmin), userService.Update())
		userRouter.GET("/list", userService.AuthMiddleware(models.PermissionAdmin), userService.List())
		userRouter.POST("/modify_pass", userService.AuthMiddleware(models.PermissionAdmin), userService.ModifyPass())
		userRouter.POST("/bind_group", userService.AuthMiddleware(models.PermissionAdmin), userService.BindGroup())

		userRouter.GET("/info", userService.AuthMiddleware(models.PermissionBase), userService.Info())
		userRouter.POST("/modify_own_pass", userService.AuthMiddleware(models.PermissionBase), userService.ModifyOwnPass())
	}

	userGroupRouter := openapiRouter.Group("/user_group", userService.AuthMiddleware(models.PermissionAdmin))
	{
		userGroupRouter.POST("/add", userGroupService.Add())
		userGroupRouter.POST("/delete", userGroupService.Delete())
		userGroupRouter.POST("/modify_name", userGroupService.ModifyName())
		userGroupRouter.POST("/batch_bind_files", userGroupService.BatchBindFiles())
		userGroupRouter.GET("/bind_files", userGroupService.GetBindFiles())
		userGroupRouter.POST("/list", userGroupService.List())
	}

	cloudTokenRouter := openapiRouter.Group("/cloud_token", userService.AuthMiddleware(models.PermissionAdmin))
	{
		cloudTokenRouter.POST("/init_qrcode", cloudTokenService.InitQrcode())
		cloudTokenRouter.POST("/check_qrcode", cloudTokenService.CheckQrcode())
		cloudTokenRouter.POST("/modify_name", cloudTokenService.ModifyName())
		cloudTokenRouter.POST("/delete", cloudTokenService.Delete())
		cloudTokenRouter.GET("/list", cloudTokenService.List())
	}

	openapiRouter.GET("/setting/get", settingService.Get())
	settingRouter := openapiRouter.Group("/setting", userService.AuthMiddleware(models.PermissionAdmin))
	{
		settingRouter.POST("/modify_name", settingService.ModifyName())
		settingRouter.POST("/refresh_key", settingService.RefreshKey())
		settingRouter.POST("/toggle_auth", settingService.ToggleAuth())
		settingRouter.POST("/toggle_local_proxy", settingService.ToggleLocalProxy())
		settingRouter.POST("/toggle_multiple_stream", settingService.ToggleMultipleStream())
		settingRouter.POST("/modify_base_url", settingService.ModifyBaseURL())
		settingRouter.POST("/toggle_enable_top_file_auto_refresh", settingService.ToggleEnableTopFileAutoRefresh())
		settingRouter.POST("/modify_job_thread_count", settingService.ModifyJobThreadCount())

		openapiRouter.POST("/setting/init_system", settingService.InitSystem())
	}

	storageRouter := openapiRouter.Group("/storage", userService.AuthMiddleware(models.PermissionAdmin))
	{
		storageRouter.POST("/add", storageService.Add())
		storageRouter.POST("/delete", storageService.Delete())
		storageRouter.POST("/modify_token", storageService.ModifyToken())
		storageRouter.GET("/list", storageService.List())

		openapiRouter.POST("/storage/deep_refresh_file", userService.AuthMiddleware(models.PermissionBase), storageService.DeepRefreshFile())
		openapiRouter.GET("/storage/file/search", userService.AuthMiddleware(models.PermissionBase), storageService.Search())
	}

	{
		openapiRouter.GET("/open_file/*path", userService.AuthMiddleware(models.PermissionBase), universalFsService.Open("/", "json"))
		davMethods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS", "PROPFIND", "MKCOL", "MOVE", "LOCK", "UNLOCK"}
		for _, method := range davMethods {
			engine.Handle(method, "/dav/*path", userService.BasicAuthMiddleware(models.PermissionDavRead), universalFsService.DavMiddleware(), universalFsService.Open("/dav", "dav"))
			engine.Handle(method, "/dav", userService.BasicAuthMiddleware(models.PermissionDavRead), universalFsService.DavMiddleware(), universalFsService.Open("/dav", "dav"))
		}
		openapiRouter.GET("/file_download", universalFsService.FileDownload())
	}

	{
		staticFS, ok := embed.StaticFS()
		if ok {
			assetsFS, _ := fs.Sub(staticFS, "assets")

			engine.StaticFS("/assets", http.FS(assetsFS))

			engine.NoRoute(func(c *gin.Context) {
				if strings.HasPrefix(c.Request.URL.Path, "/api") {
					c.Status(404)

					return
				}

				// 返回 index.html
				file, err := staticFS.Open("index.html")
				if err != nil {
					c.Status(404)

					return
				}

				defer file.Close()

				stat, _ := file.Stat()
				c.Header("Content-Type", "text/html")
				c.DataFromReader(200, stat.Size(), "text/html", file, nil)
			})
		}
	}

	logger.Info("start http server",
		zap.Int("port", config.Port),
	)

	return engine.Run(fmt.Sprintf(":%d", config.Port))
}
