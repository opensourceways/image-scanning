package app

import (
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/utils"
)

const (
	script = "./trivy_env.sh"
)

type TrivyService interface {
	InitTrivyEnv() error
	UpdateTrivyDB()
}

func NewTrivyService(r *TrivyRepo) *trivyService {
	return &trivyService{
		repo: r,
	}
}

type trivyService struct {
	repo *TrivyRepo
}

func (t *trivyService) InitTrivyEnv() error {
	out, err := utils.RunCmd(script, "init", t.repo.Trivy, t.repo.TrivyDB, t.repo.VulnList)
	if err != nil {
		logrus.Errorf("init trivy env failed: %s,output: %s", err.Error(), out)
		return err
	}

	return nil
}

func (t *trivyService) UpdateTrivyDB() {
	if out, err := utils.RunCmd(script, "update"); err != nil {
		logrus.Errorf("init trivy env failed: %s,output: %s", err.Error(), out)
	}
}
