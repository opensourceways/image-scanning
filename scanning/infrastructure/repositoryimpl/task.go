package repositoryimpl

import (
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/common/infrastructure/postgresql"
	"github.com/opensourceways/image-scanning/scanning/domain"
)

func NewTaskImpl() *taskImpl {
	do := &TaskDO{}
	if err := postgresql.DB().AutoMigrate(do); err != nil {
		logrus.Fatalf("auto migrate table %s failed: %v", do.TableName(), err)
	}

	return &taskImpl{
		Impl: postgresql.DAO(do.TableName()),
	}
}

type taskImpl struct {
	postgresql.Impl
}

func (impl *taskImpl) Save(task domain.Task) error {
	do := ToTaskDO(task)

	return impl.DB().Save(&do).Error
}

func (impl *taskImpl) Find(task domain.Task) (domain.Task, error) {
	do := TaskDO{}
	if task.Community != "" {
		do.Community = task.Community
	}

	if task.Registry != nil {
		do.Registry = task.Registry.String()
	}

	if task.Namespace != "" {
		do.Namespace = task.Namespace
	}

	if task.Image != "" {
		do.Image = task.Image
	}

	if task.Tag != "" {
		do.Tag = task.Tag
	}

	if err := impl.DB().First(&do, &do).Error; err != nil {
		return domain.Task{}, err
	}

	return do.ToTask(), nil
}

func (impl *taskImpl) FindAll(name string) ([]domain.Task, error) {
	var dos []TaskDO
	if err := impl.DB().Order(fieldId).Where(TaskDO{Community: name}).Find(&dos).Error; err != nil {
		return nil, err
	}

	tasks := make([]domain.Task, len(dos))
	for i := range dos {
		tasks[i] = dos[i].ToTask()
	}

	return tasks, nil
}

func (impl *taskImpl) DeleteByIds(ids []int64) error {
	return impl.DB().Delete(&TaskDO{}, ids).Error
}
