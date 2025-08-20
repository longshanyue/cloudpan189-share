package shared

import (
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/messagebus"
	"go.uber.org/zap"
)

var (
	FileBus  = messagebus.New(messagebus.DefaultConfig(), zap.NewNop())
	MediaBus = messagebus.New(messagebus.DefaultConfig(), zap.NewNop())
)
