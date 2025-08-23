package utils

import (
	"crypto/md5"
	"fmt"
)

func MD5(data []byte) string {
	hash := md5.Sum(data)

	return fmt.Sprintf("%x", hash)
}
