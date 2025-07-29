package app

import (
	"os/exec"

	"github.com/sirupsen/logrus"
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
	cmd := exec.Command(script, "init", t.repo.Trivy, t.repo.TrivyDB, t.repo.VulnList)
	if out, err := cmd.Output(); err != nil {
		logrus.Errorf("init trivy env failed: %s,output: %s", err.Error(), string(out))
		return err
	}

	return nil
}

func (t *trivyService) UpdateTrivyDB() {
	cmd := exec.Command(script, "update")
	if out, err := cmd.Output(); err != nil {
		logrus.Errorf("init trivy env failed: %s,output: %s", err.Error(), string(out))
	}
}
