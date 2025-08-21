package bus

import "github.com/xxcheng123/cloudpan189-share/internal/models"

const (
	TopicFileRefreshFile      = "file::refresh::file"
	TopicFileDeleteFile       = "file::delete::file"
	TopicFileScanTop          = "file::scan::top"
	TopicFileRebuildMediaFile = "file::rebuild::media::file"
)

type TopicFileRefreshRequest struct {
	FID  int64 `json:"fid"`
	Deep bool  `json:"deep"`
}

type TopicFileDeleteRequest struct {
	FID int64 `json:"fid"`
}

type TopicFileRebuildMediaFileRequest struct {
	MediaTypes []models.MediaType
}

const (
	TopicMediaAddStrmFile    = "media::add::strm::file"
	TopicMediaDeleteLinkFile = "media::delete::link::file"
	TopicMediaClearEmptyDir  = "media::clear::empty::dir"
	TopicMediaClearAllMedia  = "media::clear::all::media"
)

type TopicMediaAddStrmFileRequest struct {
	FileID int64  `json:"fileId"`
	Path   string `json:"path"`
}

type TopicMediaDeleteLinkFileRequest struct {
	FileID int64 `json:"fileId"`
}

type TopicMediaClearEmptyDirRequest struct {
}

type TopicMediaClearAllMediaRequest struct {
	MediaTypes []models.MediaType
}
