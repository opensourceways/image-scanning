package app

import (
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/scanning/domain/platform"
	"github.com/opensourceways/image-scanning/scanning/domain/repository"
	"github.com/opensourceways/image-scanning/scanning/infrastructure/platformimpl"
)

var (
	handlers      map[string]*communityHandler
	scanConfigSha map[string]string
)

type TaskService interface {
	GenerateTask()
	ExecTask()
}

func NewTaskService(cs []platform.Community, repo repository.Task) *taskService {
	return &taskService{
		communities: cs,
		repo:        repo,
	}
}

type taskService struct {
	repo        repository.Task
	communities []platform.Community
}

func (t *taskService) GenerateTask() {
	for _, c := range t.communities {
		var ptat platform.Platform
		if c.IsGiteePlatform() {
			ptat = platformimpl.NewGiteeImpl(&c)
		} else {
			// 如果需要支持其他平台在此处扩展优化
			logrus.Errorf("unsupported platform %v", c)
			continue
		}

		scanConfig, sha, err := ptat.DownloadScanConfig()
		if err != nil {
			logrus.Errorf("get scan config of %s failed: %s", c.Name, err.Error())
			continue
		}

		if t.shaCheckNotChange(c.Name, sha) {
			logrus.Infof("sha of %s not change", c.Name)
			continue
		}

		ptat.SetOutput(scanConfig.Scanner.Global.Output)
		handler := newCommunityHandler(c, t.repo, ptat)
		handler.generateTask(scanConfig)

		if len(handlers) == 0 {
			handlers = make(map[string]*communityHandler)
		}

		handlers[c.Name] = handler
	}
}

func (t *taskService) shaCheckNotChange(communityName, newSha string) bool {
	if len(scanConfigSha) == 0 {
		scanConfigSha = make(map[string]string)
	}

	oldSha, ok := scanConfigSha[communityName]
	scanConfigSha[communityName] = newSha
	if !ok {
		return false
	}

	return oldSha == newSha
}

func (t *taskService) ExecTask() {
	for _, handler := range handlers {
		tasks, err := handler.repo.FindAll(handler.name)
		if err != nil {
			logrus.Errorf("find all tasks of %s failed when exec task: %s", handler.name, err.Error())
			continue
		}

		for _, task := range tasks {
			if !task.IsNeedToScan() {
				continue
			}

			if err = handler.handleTask(&task); err != nil {
				logrus.Errorf("handle task %s failed: %s", task.UniqueKey(), err.Error())
				continue
			}

			task.UpdateLastScanTime()
			if err = handler.repo.Save(task); err != nil {
				logrus.Errorf("save task %s when exec failed: %s", task.UniqueKey(), err.Error())
			}
		}
	}
}
