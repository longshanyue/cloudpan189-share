package bus

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/enc"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

func generateDownloadURLWithNeverExpire(fid int64) string {
	baseURL := ""
	if shared.StrmBaseURL != "" {
		baseURL = shared.StrmBaseURL
	} else if shared.Setting != nil {
		baseURL = shared.Setting.BaseURL
	}

	values := enc.Enc(url.Values{
		"id":        []string{fmt.Sprintf("%d", fid)},
		"random":    []string{uuid.NewString()},
		"timestamp": []string{"-1"},
	}, shared.Setting.SaltKey)

	return fmt.Sprintf("%s/api/file_download?%s", baseURL, values.Encode())
}
