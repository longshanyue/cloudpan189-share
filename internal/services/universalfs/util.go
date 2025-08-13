package universalfs

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
