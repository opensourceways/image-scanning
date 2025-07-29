package platform

import "github.com/opensourceways/image-scanning/scanning/domain"

type Community struct {
	Name               string   `json:"name"                 required:"true"`
	GiteeToken         string   `json:"gitee_token"`
	ScanConfigLocation Location `json:"scan_config_location" required:"true"`
}

func (c Community) IsGiteePlatform() bool {
	return c.GiteeToken != ""
}

type Location struct {
	Ref  string `json:"ref"  required:"true"`
	Repo string `json:"repo" required:"true"`
	Path string `json:"path" required:"true"`
}

type Platform interface {
	Upload(string, string) error
	SetOutput(output domain.Output)
	DownloadScanConfig() (domain.ScanConfig, string, error)
}
