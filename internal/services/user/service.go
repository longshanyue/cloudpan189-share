package user

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	Add() gin.HandlerFunc
	Del() gin.HandlerFunc
	Update() gin.HandlerFunc
	List() gin.HandlerFunc
	Login() gin.HandlerFunc
	RefreshToken() gin.HandlerFunc
	AuthMiddleware(permission uint8) gin.HandlerFunc
	BasicAuthMiddleware(permission uint8) gin.HandlerFunc
	Info() gin.HandlerFunc
	ModifyPass() gin.HandlerFunc
	ModifyOwnPass() gin.HandlerFunc
}

type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:     db,
		logger: logger,
	}
}

// 生成 md5
func hash(input string) string {
	data := []byte(input)
	has := md5.Sum(data)

	return fmt.Sprintf("%x", has)
}

// Claims JWT令牌结构体
type Claims struct {
	UserId      int64  `json:"user_id"`
	Username    string `json:"username"`
	UserVersion int    `json:"user_version"`
	jwt.RegisteredClaims
}

// Token过期时间
const (
	AccessTokenExpire  = time.Hour * 24     // 24小时
	RefreshTokenExpire = time.Hour * 24 * 7 // 7天
)

// generateAccessToken 生成访问Token
func (s *service) generateAccessToken(userId int64, username string, userVersion int) (string, error) {
	claims := &Claims{
		UserId:      userId,
		Username:    username,
		UserVersion: userVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   "access-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(shared.Setting.SaltKey))
}

// generateRefreshToken 生成刷新Token
func (s *service) generateRefreshToken(userId int64, username string, userVersion int) (string, error) {
	claims := &Claims{
		UserId:      userId,
		Username:    username,
		UserVersion: userVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   "refresh-token",
			ID:        fmt.Sprintf("%d", userId),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(shared.Setting.SaltKey))
}

// ParseAccessToken 解析访问Token
func (s *service) ParseAccessToken(tokenString string) (int64, string, int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(shared.Setting.SaltKey), nil
	})

	if err != nil {
		return 0, "", 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Subject != "access-token" {
			return 0, "", 0, fmt.Errorf("invalid access token subject")
		}

		return claims.UserId, claims.Issuer, claims.UserVersion, nil
	}

	return 0, "", 0, fmt.Errorf("invalid access token")
}

// parseRefreshToken 解析刷新Token
func (s *service) parseRefreshToken(tokenString string) (int64, string, int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(shared.Setting.SaltKey), nil
	})

	if err != nil {
		return 0, "", 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Subject != "refresh-token" {
			return 0, "", 0, fmt.Errorf("invalid refresh token subject")
		}

		return claims.UserId, claims.Issuer, claims.UserVersion, nil
	}

	return 0, "", 0, fmt.Errorf("invalid refresh token")
}
