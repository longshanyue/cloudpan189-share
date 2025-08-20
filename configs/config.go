package configs

import (
	"path/filepath"
)

type Config struct {
	Port    int    `json:"port,default=12395"`
	DBFile  string `json:"dbFile,default=data/share.db"`
	LogFile string `json:"logFile,default=logs/share.log"`
	// Deprecated: use MediaDir instead.
	FileDir  string `json:"fileDir,default=datadir"`
	MediaDir string `json:"mediaDir,default=media_dir"`
}

func (c *Config) MediaJoinPath(paths ...string) string {
	tp := []string{c.MediaDir}
	tp = append(tp, paths...)

	return filepath.Join(tp...)
}
