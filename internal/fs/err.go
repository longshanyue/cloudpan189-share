package fs

import "github.com/pkg/errors"

var (
	FileNil                    = errors.New("file can not be nil")
	RootDirProhibitsCreateFile = errors.New("root dir can not create file")
	WhereParamInvalid          = errors.New("where param invalid")
)
