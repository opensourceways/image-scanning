package platformimpl

import (
	"encoding/base64"
	"path"
	"strings"

	"github.com/opensourceways/robot-github-lib/client"
	"sigs.k8s.io/yaml"

	"github.com/opensourceways/image-scanning/scanning/domain"
)

const (
	uploadDefaultBranchOfGithub = "main"
)

func NewGithubImpl(c *domain.Community) *githubImpl {
	githubClient := client.NewClient(func() []byte {
		return []byte(c.Token)
	})

	return &githubImpl{
		client:    githubClient,
		community: c,
	}
}

type githubImpl struct {
	client    client.Client
	output    domain.Output
	community *domain.Community
}

func (impl *githubImpl) SetOutput(output domain.Output) {
	impl.output = output
}

func (impl *githubImpl) DownloadScanConfig() (scanConfig domain.ScanConfig, sha string, err error) {
	scl := impl.community.ScanConfigLocation
	content, err := impl.client.GetPathContent(impl.community.Name, scl.Repo, scl.Path, scl.Ref)
	if err != nil {
		return
	}

	sha = *content.SHA

	cleanedContent := strings.ReplaceAll(*content.Content, "\n", "")
	decodeData, err := base64.StdEncoding.DecodeString(cleanedContent)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(decodeData, &scanConfig)

	return
}

func (impl *githubImpl) Upload(content, mdPath string) error {
	repoName := impl.output.GetRepoName()
	filePath := path.Join(impl.output.Path, mdPath)

	var sha string
	fileContent, err := impl.client.GetPathContent(impl.community.Name, repoName, filePath, uploadDefaultBranchOfGithub)
	if err != nil {
		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	} else {
		sha = *fileContent.SHA
	}

	return impl.client.CreateFile(impl.community.Name, repoName, filePath,
		uploadDefaultBranchOfGithub, uploadDefaultCommitMsg, sha, []byte(content),
	)
}
