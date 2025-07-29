package repository

import "github.com/opensourceways/image-scanning/scanning/domain"

type Task interface {
	Save(task domain.Task) error
	Find(task domain.Task) (domain.Task, error)
	FindAll(name string) (tasks []domain.Task, err error)
	DeleteByIds(ids []int64) error
}
