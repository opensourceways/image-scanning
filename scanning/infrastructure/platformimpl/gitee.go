package platformimpl

import (
	"encoding/base64"
	"path"
	"strings"

	"github.com/opensourceways/robot-gitee-lib/client"
	"sigs.k8s.io/yaml"

	"github.com/opensourceways/image-scanning/scanning/domain"
	"github.com/opensourceways/image-scanning/scanning/domain/platform"
)

const (
	uploadDefaultBranch    = "master"
	uploadDefaultCommitMsg = "image scanning result"
)

func NewGiteeImpl(c *platform.Community) *giteeImpl {
	giteeClient := client.NewClient(func() []byte {
		return []byte(c.GiteeToken)
	})

	return &giteeImpl{
		client:    giteeClient,
		community: c,
	}
}

type giteeImpl struct {
	client    client.Client
	output    domain.Output
	community *platform.Community
}

func (impl *giteeImpl) SetOutput(output domain.Output) {
	impl.output = output
}

func (impl *giteeImpl) DownloadScanConfig() (scanConfig domain.ScanConfig, sha string, err error) {
	scl := impl.community.ScanConfigLocation
	content, err := impl.client.GetPathContent(impl.community.Name, scl.Repo, scl.Path, scl.Ref)
	if err != nil {
		return
	}

	sha = content.Sha

	decodeData, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(decodeData, &scanConfig)

	return
}

func (impl *giteeImpl) Upload(content, mdPath string) error {
	var fileIsNotExist bool
	filePath := path.Join(impl.output.Path, mdPath)
	fileContent, err := impl.client.GetPathContent(impl.community.Name, impl.output.Repo, filePath, "master")
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") {
			fileIsNotExist = true
		} else {
			return err
		}
	}

	if fileIsNotExist {
		_, err = impl.client.CreateFile(impl.community.Name, impl.output.Repo,
			uploadDefaultBranch, filePath, content, uploadDefaultCommitMsg)
	} else {
		_, err = impl.client.UpdateFile(impl.community.Name, impl.output.Repo,
			uploadDefaultBranch, filePath, content, fileContent.Sha, uploadDefaultCommitMsg)
	}

	return err
}
