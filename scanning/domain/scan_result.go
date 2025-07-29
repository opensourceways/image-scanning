package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	securityBulletinUrlPrefix = "https://www.openeuler.org/zh/security/security-bulletins/detail/?id="
)

type ScanResult struct {
	CreatedAt string   `json:"createdAt"`
	Metadata  Metadata `json:"Metadata"`
	Results   []Result `json:"Results"`
}

type Metadata struct {
	RepoTags    []string    `json:"RepoTags"`
	ImageConfig ImageConfig `json:"ImageConfig"`
}

type ImageConfig struct {
	OS   string `json:"os"`
	Arch string `json:"architecture"`
}

func (i ImageConfig) GenArch() string {
	return fmt.Sprintf("%s/%s", i.OS, i.Arch)
}

type Result struct {
	Target          string          `json:"Target"`
	Class           string          `json:"Class"`
	Type            string          `json:"Type"`
	Vulnerabilities []Vulnerability `json:"Vulnerabilities"`
}

func (r Result) isValid() bool {
	return r.Class == "os-pkgs" && r.Type == "openEuler"
}

type Vulnerability struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Status           string `json:"Status"`
	Severity         string `json:"Severity"`
}

func (v Vulnerability) securityBulletinUrl() string {
	return securityBulletinUrlPrefix + v.VulnerabilityID
}

func (v Vulnerability) securityBulletinUrlOfMarkdown() string {
	return fmt.Sprintf("[%s](%s)", v.VulnerabilityID, v.securityBulletinUrl())
}

func (r ScanResult) ToMarkdown() string {
	tableHead :=
		`|  软件包  | 安全公告 | 严重级别 |  状态  | 安装版本 | 修复版本 |
| :----- | :-----  | :-----  | :----- | :----- | :----- |`

	rowFormat := `| %s | %s |  %s |  %s |  %s |  %s |`

	var tableBody []string
	for _, result := range r.Results {
		if !result.isValid() {
			continue
		}

		for _, vuln := range result.Vulnerabilities {
			row := fmt.Sprintf(rowFormat,
				vuln.PkgName,
				vuln.securityBulletinUrlOfMarkdown(),
				vuln.Severity,
				vuln.Status,
				vuln.InstalledVersion,
				vuln.FixedVersion,
			)

			tableBody = append(tableBody, row)
		}
	}

	var scanResult string
	if len(tableBody) > 0 {
		scanResult = tableHead + "\n" + strings.Join(tableBody, "\n")
	} else {
		scanResult = "无漏洞"
	}

	return scanResult + "\n"
}

type ArchResult struct {
	Err        error
	ScanResult ScanResult
}

func BuildContent(ars map[string]ArchResult) string {
	content := fmt.Sprintf("# 扫描时间：%s\n", time.Now().Format(time.DateTime))
	for arch, ar := range ars {
		content += fmt.Sprintf("--- \n ### 扫描架构：%s \n", arch)

		if ar.Err == nil {
			content += ar.ScanResult.ToMarkdown()
		} else {
			content += ar.Err.Error() + "\n"
		}
	}

	return content
}
