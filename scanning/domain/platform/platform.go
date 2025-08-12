package platform

import "github.com/opensourceways/image-scanning/scanning/domain"

type Platform interface {
	Upload(string, string) error
	SetOutput(output domain.Output)
	DownloadScanConfig() (domain.ScanConfig, string, error)
}
