package shared

//go:generate go run ../../cmd/generate_setting/main.go

import "github.com/xxcheng123/cloudpan189-share/internal/models"

// Setting 实例保留在这里，因为它不需要生成
var Setting = &models.Setting{}
