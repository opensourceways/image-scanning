package app

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/scanning/domain"
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

func NewTaskService(cs []domain.Community, con Concurrency, repo repository.Task) *taskService {
	return &taskService{
		communities: cs,
		repo:        repo,
		taskChan:    make(chan domain.Task, 1000),
		concurrency: con,
	}
}

type taskService struct {
	repo        repository.Task
	communities []domain.Community

	mu          sync.Mutex
	taskChan    chan domain.Task
	concurrency Concurrency
}

func (t *taskService) getPlatform(c *domain.Community) platform.Platform {
	switch c.Platform {
	case domain.PlatformGitee:
		return platformimpl.NewGiteeImpl(c)
	case domain.PlatformGithub:
		return platformimpl.NewGithubImpl(c)
	default:
		return nil
	}
}

func (t *taskService) GenerateTask() {
	for _, c := range t.communities {
		plat := t.getPlatform(&c)
		if plat == nil {
			logrus.Errorf("unsupported platform %v", c)
			continue
		}

		scanConfig, sha, err := plat.DownloadScanConfig()
		if err != nil {
			logrus.Errorf("get scan config of %s failed: %s", c.Name, err.Error())
			continue
		}

		if t.shaCheckNotChange(c.Name, sha) {
			logrus.Infof("sha of %s not change", c.Name)
			continue
		}

		plat.SetOutput(scanConfig.Scanner.Global.Output)
		handler := newCommunityHandler(c, t.repo, plat)
		handler.generateTask(scanConfig)

		if len(handlers) == 0 {
			handlers = make(map[string]*communityHandler)
		}

		t.mu.Lock()
		handlers[c.Name] = handler
		t.mu.Unlock()
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
	// 首次执行任务或者配置文件新增任务时，拉取镜像十分费时，可能会跨越到下次定时任务执行的时间，
	// 如果channel还有未执行完的任务，则跳过，避免重复执行
	if len(t.taskChan) > 0 {
		return
	}

	go t.loadTask()

	t.handleTaskConcurrently()
}

func (t *taskService) loadTask() {
	defer t.recovery()

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

			t.taskChan <- task
		}
	}
}

func (t *taskService) handleTaskConcurrently() {
	for i := 1; i <= t.concurrency.Num; i++ {
		go func() {
			defer t.recovery()

			for task := range t.taskChan {
				handler, ok := handlers[task.Community]
				if !ok {
					continue
				}

				if err := handler.handleTask(&task); err != nil {
					logrus.Errorf("handle task %s failed: %s", task.UniqueKey(), err.Error())
					continue
				}

				task.UpdateLastScanTime()
				if err := handler.repo.Save(task); err != nil {
					logrus.Errorf("save task %s when exec failed: %s", task.UniqueKey(), err.Error())
				}
			}
		}()
	}
}

func (t *taskService) recovery() {
	if r := recover(); r != nil {
		logrus.Errorf("exec task panic %v", r)
	}
}
