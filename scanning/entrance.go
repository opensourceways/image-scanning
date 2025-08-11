package scanning

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/config"
	"github.com/opensourceways/image-scanning/scanning/app"
	"github.com/opensourceways/image-scanning/scanning/infrastructure/repositoryimpl"
)

var instance *scanner

type scanner struct {
	job          *cron.Cron
	cfg          *config.Config
	trivyService app.TrivyService
	taskService  app.TaskService
}

func Run(cfg *config.Config) {
	trivyService := app.NewTrivyService(&cfg.TrivyRepo)
	taskService := app.NewTaskService(cfg.Community, repositoryimpl.NewTaskImpl())

	instance = &scanner{
		job:          cron.New(),
		cfg:          cfg,
		trivyService: trivyService,
		taskService:  taskService,
	}

	if err := instance.trivyService.InitTrivyEnv(); err != nil {
		logrus.Fatalf("init trivy env failed: %s", err.Error())
	}

	// 程序启动先同步一次任务
	instance.taskService.GenerateTask()

	instance.addJob()

	instance.job.Run()
}

func (s *scanner) addJob() {
	// 每小时同步一次配置文件，更新扫描任务
	if _, err := s.job.AddFunc("55 * * * *", s.taskService.GenerateTask); err != nil {
		logrus.Fatalf("add cron job [GenerateTask]  failed: %s", err.Error())
	}

	// 看配置要求调整执行任务的粒度，保证覆盖就可以，一般不会太频繁
	if _, err := s.job.AddFunc("*/10 * * * *", s.taskService.ExecTask); err != nil {
		logrus.Fatalf("add cron job [ExecTask]  failed: %s", err.Error())
	}

	// 由于openeuler每周发布一次公告，故每周更新第一次即可
	if _, err := s.job.AddFunc("0 1 * * 0", s.trivyService.UpdateTrivyDB); err != nil {
		logrus.Fatalf("add cron job [UpdateTrivyDB]  failed: %s", err.Error())
	}
}
