package configs

type Config struct {
	Port    int    `json:"port,default=12395"`
	DBFile  string `json:"dbFile,default=data/share.db"`
	LogFile string `json:"logFile,default=logs/share.log"`
}
