package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/samber/lo"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AutoLoginJob struct {
	db     *gorm.DB
	mu     sync.Mutex
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewAutoLoginJob(db *gorm.DB, logger *zap.Logger) Job {
	return &AutoLoginJob{
		db:     db,
		logger: logger.With(zap.String("job", "auto_login")),
	}
}

func (s *AutoLoginJob) Start(ctx context.Context) error {
	if !s.mu.TryLock() {
		return ErrJobRunning
	}

	defer s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(ctx)

	gopool.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				s.logger.Info("auto login job stopped")

				return
			case <-time.After(time.Hour):
			}

			// 查询过期时间还剩7天的 token
			var tokens = make([]*models.CloudToken, 0)
			if err := s.db.WithContext(ctx).Where("login_type = ?", models.LoginTypePassword).Where("expires_in < ?", time.Now().Unix()+7*24*3600).Find(&tokens).Error; err != nil {
				s.logger.Error("query cloud token error", zap.Error(err))

				continue
			}

			retryTimesMap := make(map[int64]int)

			tokens = lo.Filter(tokens, func(token *models.CloudToken, index int) bool {
				val, err := utils.Int(token.Addition[models.CloudTokenAdditionAutoLoginTimes])
				if err != nil {
					val = 0 // 如果获取失败，默认为0次重试
				}

				retryTimesMap[token.ID] = val

				return val < 3
			})

			s.logger.Info("auto login job started", zap.Int("count", len(tokens)))

			// 执行刷新
			for _, token := range tokens {
				loginResult, loginErr := cloudpan.AppLogin(token.Username, token.Password)

				updateMap := make(map[string]interface{})

				if loginErr != nil {
					s.logger.Error("auto login error", zap.Error(loginErr), zap.Any("token", token))
					retryTimesMap[token.ID]++

					token.Addition[models.CloudTokenAdditionAutoLoginResultKey] = fmt.Sprintf("%s，刷新 token 失败。%s", time.Now().Format(time.DateTime), loginErr.Err)
					token.Addition[models.CloudTokenAdditionAutoLoginTimes] = retryTimesMap[token.ID]

					updateMap["addition"] = token.Addition
				} else {
					s.logger.Info("auto login success", zap.Any("token", token), zap.Any("loginResult", loginResult))

					token.Addition[models.CloudTokenAdditionAutoLoginResultKey] = fmt.Sprintf("%s，刷新 token 成功。", time.Now().Format(time.DateTime))
					token.Addition[models.CloudTokenAdditionAutoLoginTimes] = 0

					updateMap["addition"] = token.Addition
					updateMap["expires_in"] = loginResult.SskAccessTokenExpiresIn
					updateMap["access_token"] = loginResult.SskAccessToken
				}

				if err := s.db.WithContext(ctx).Model(&models.CloudToken{}).Where("id = ?", token.ID).Updates(updateMap).Error; err != nil {
					s.logger.Error("update cloud token error", zap.Error(err))
				}
			}
		}
	})

	return nil
}

func (s *AutoLoginJob) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
