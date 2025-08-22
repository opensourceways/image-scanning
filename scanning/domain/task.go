package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/opensourceways/image-scanning/scanning/domain/primitive"
)

const (
	ImagesDir = "images"
)

type Task struct {
	Id           int64
	Community    string
	Registry     primitive.Registry
	Namespace    string
	Image        string
	Tag          string
	Arch         []string
	Interval     int
	LastScanTime time.Time
}

func GenerateTask(communityName string, cfg *ScanConfig) map[string]Task {
	initGlobalConfig(communityName, &cfg.Scanner.Global)

	taskSets := make(map[string]Task)
	for _, repo := range cfg.Repos {
		repo.genTask(communityName, taskSets)
	}

	for _, image := range cfg.Images {
		image.genTask(communityName, taskSets)
	}

	return taskSets
}

func (t *Task) UniqueKey() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", t.Community, t.Registry, t.Namespace, t.Image, t.Tag)
}

func (t *Task) UpdateIntervalAndArch(interval int, arch []string) {
	t.Interval = interval
	t.Arch = arch
}

func (t *Task) IsNeedToScan() bool {
	if t.LastScanTime.IsZero() {
		return true
	}

	nextScanTime := t.LastScanTime.Add(time.Second * time.Duration(t.Interval))

	return time.Now().After(nextScanTime)
}

func (t *Task) ImagePath() string {
	return fmt.Sprintf("%s/%s/%s:%s", t.Registry, t.Namespace, t.Image, t.Tag)
}

func (t *Task) LocalImagePath(arch string) string {
	return fmt.Sprintf("%s/%s_%s_%s_%s_%s", ImagesDir, t.Registry, t.Namespace, t.Image, t.Tag, arch)
}

func (t *Task) UpdateLastScanTime() {
	t.LastScanTime = time.Now()
}

func (t *Task) MarkdownPath() string {
	return fmt.Sprintf("%s/%s/%s/%s.md", t.Registry, t.Namespace, t.Image, t.Tag)
}

func (t *Task) FormatArch() []string {
	var formatArch []string
	for _, v := range t.Arch {
		formatArch = append(formatArch, strings.TrimPrefix(v, "linux/"))
	}

	return formatArch
}
