package utils

import (
	"net/http"
)

var NoFollowRedirectHttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // 停止跟随重定向
	},
}
