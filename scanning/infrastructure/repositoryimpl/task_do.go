package repositoryimpl

import (
	"strings"
	"time"

	"github.com/opensourceways/image-scanning/scanning/domain"
	"github.com/opensourceways/image-scanning/scanning/domain/primitive"
)

const (
	fieldId = "id"
)

type TaskDO struct {
	Id           int64     `gorm:"column:id;primaryKey; autoIncrement"`
	Community    string    `gorm:"column:community;comment:社区"`
	Registry     string    `gorm:"column:registry;comment:镜像站"`
	Namespace    string    `gorm:"column:namespace;comment:镜像站命名空间"`
	Image        string    `gorm:"column:image;comment:镜像名"`
	Tag          string    `gorm:"column:tag;comment:镜像tag"`
	Arch         string    `gorm:"column:arch;comment:架构"`
	Interval     int       `gorm:"column:interval;comment:扫描间隔，单位秒"`
	LastScanTime time.Time `gorm:"column:last_scan_time;comment:上次扫描时间"`
	CreatedAt    time.Time `gorm:"column:created_at;<-:create"`
	UpdatedAt    time.Time `gorm:"column:updated_at;<-:update"`
}

func (do *TaskDO) TableName() string {
	return "task"
}

func ToTaskDO(task domain.Task) TaskDO {
	return TaskDO{
		Id:           task.Id,
		Community:    task.Community,
		Registry:     task.Registry.String(),
		Namespace:    task.Namespace,
		Image:        task.Image,
		Tag:          task.Tag,
		Arch:         strings.Join(task.Arch, ","),
		Interval:     task.Interval,
		LastScanTime: task.LastScanTime,
	}
}

func (do *TaskDO) ToTask() domain.Task {
	return domain.Task{
		Id:           do.Id,
		Community:    do.Community,
		Registry:     primitive.CreateRegistry(do.Registry),
		Namespace:    do.Namespace,
		Image:        do.Image,
		Tag:          do.Tag,
		Arch:         strings.Split(do.Arch, ","),
		Interval:     do.Interval,
		LastScanTime: do.LastScanTime,
	}
}
