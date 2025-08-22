package bus

import "github.com/xxcheng123/cloudpan189-share/internal/models"

type TopicType string

func (t TopicType) String() string {
	return string(t)
}

const (
	TopicFileRefreshFile      = "topic::file::refresh::file"
	TopicFileDeleteFile       = "topic::file::delete::file"
	TopicFileScanTop          = "topic::file::scan::top"
	TopicFileRebuildMediaFile = "file::rebuild::media::file"

	TopicMediaAddStrmFile    = "topic::media::add::strm::file"
	TopicMediaDeleteLinkFile = "topic::media::delete::link::file"
	TopicMediaClearEmptyDir  = "topic::media::clear::empty::dir"
	TopicMediaClearAllMedia  = "topic::media::clear::all::media"
)

type TopicFileRefreshFileRequest struct {
	FileId int64 `json:"fileId"`
	Deep   bool  `json:"deep"`
}

type TopicFileDeleteRequest struct {
	FileId int64 `json:"fid"`
}

type TopicFileRebuildMediaFileRequest struct {
	MediaTypes []models.MediaType
}

type TopicMediaDeleteLinkFileRequest struct {
	FileId int64 `json:"fileId"`
}

type TopicMediaClearEmptyDirRequest struct {
}

type TopicMediaAddStrmFileRequest struct {
	FileID int64  `json:"fileId"`
	Path   string `json:"path"`
}

type TopicMediaClearAllMediaRequest struct {
	MediaTypes []models.MediaType
}
