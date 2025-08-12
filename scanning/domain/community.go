package domain

const (
	PlatformGitee  = "gitee"
	PlatformGithub = "github"
)

type Community struct {
	Name               string   `json:"name"                 required:"true"`
	Token              string   `json:"token"                required:"true"`
	Platform           string   `json:"platform"             required:"true"`
	ScanConfigLocation Location `json:"scan_config_location" required:"true"`
}

type Location struct {
	Ref  string `json:"ref"  required:"true"`
	Repo string `json:"repo" required:"true"`
	Path string `json:"path" required:"true"`
}
