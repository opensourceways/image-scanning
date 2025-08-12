package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/opensourceways/image-scanning/scanning/domain"
	"github.com/opensourceways/image-scanning/scanning/domain/platform"
	"github.com/opensourceways/image-scanning/scanning/domain/repository"
)

func newCommunityHandler(c domain.Community, repo repository.Task, p platform.Platform) *communityHandler {
	return &communityHandler{
		name:     c.Name,
		repo:     repo,
		platform: p,
	}
}

type communityHandler struct {
	name     string
	repo     repository.Task
	platform platform.Platform
}

func (h *communityHandler) generateTask(scanConfig domain.ScanConfig) {
	taskSets := domain.GenerateTask(h.name, &scanConfig)
	if err := h.clearOldTasks(taskSets); err != nil {
		logrus.Errorf("clear old task of %s failed: %s", h.name, err.Error())
	}

	for _, task := range taskSets {
		if err := h.saveTask(task); err != nil {
			logrus.Errorf("save task failed: %s", err.Error())
		}
	}
}

// clearOldTasks 清理配置中没有的任务不能直接删表再重建，这样会丢失上次执行时间，还会影响将要执行的任务
// 只能找出配置中不存在的任务进行指定清理
func (h *communityHandler) clearOldTasks(newTasks map[string]domain.Task) error {
	oldTasks, err := h.repo.FindAll(h.name)
	if err != nil {
		return err
	}

	var clearIds []int64
	for _, oldTask := range oldTasks {
		_, ok := newTasks[oldTask.UniqueKey()]
		if !ok {
			clearIds = append(clearIds, oldTask.Id)
		}
	}

	if len(clearIds) == 0 {
		return nil
	}

	return h.repo.DeleteByIds(clearIds)
}

func (h *communityHandler) saveTask(newTask domain.Task) error {
	oldTask, err := h.repo.Find(newTask)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			oldTask = newTask
		} else {
			return err
		}
	} else {
		oldTask.UpdateIntervalAndArch(newTask.Interval, newTask.Arch)
	}

	return h.repo.Save(oldTask)
}

func (h *communityHandler) handleTask(task *domain.Task) error {
	ars := make(map[string]domain.ArchResult, len(task.Arch))
	for _, arch := range task.Arch {
		param := []string{
			"image",
			"--quiet",
			"--skip-db-update",
			"-f", "json",
			"--scanners", "vuln",
			"--cache-dir", "./trivy_resource/",
			"--platform", arch,
			task.ImagePath(),
		}

		ars[arch] = h.handleArch(param)
	}

	return h.platform.Upload(domain.BuildContent(ars), task.MarkdownPath())
}

func (h *communityHandler) handleArch(param []string) domain.ArchResult {
	var ar domain.ArchResult
	var out, stderr bytes.Buffer

	cmd := exec.Command("./trivy", param...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		ar.Err = errors.New(stderr.String())
	} else {
		ar.Err = json.Unmarshal(out.Bytes(), &ar.ScanResult)
	}

	return ar
}
