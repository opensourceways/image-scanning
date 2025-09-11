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
	taskService := app.NewTaskService(cfg.Community, cfg.Concurrency, repositoryimpl.NewTaskImpl())

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

	// 每6小时，trivy的标准周期
	if _, err := s.job.AddFunc("0 */6 * * *", s.trivyService.UpdateTrivyDB); err != nil {
		logrus.Fatalf("add cron job [UpdateTrivyDB]  failed: %s", err.Error())
	}

	// 镜像站的镜像会按照月级更新，定期清理重新拉取
	if _, err := s.job.AddFunc("30 14 11 * *", s.taskService.ClearImages); err != nil {
		logrus.Fatalf("add cron job [ClearImages]  failed: %s", err.Error())
	}
}
