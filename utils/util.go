package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"sigs.k8s.io/yaml"
)

// LoadFromYaml reads a YAML file from the given path and unmarshals it into the provided interface.
func LoadFromYaml(path string, cfg interface{}) error {
	b, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, cfg)
}

// StringToInterval s support format: 24h,1d,1w,30m
func StringToInterval(s string) (int, error) {
	length := len(s)
	if length < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	num := s[:length-1]
	unit := s[length-1]

	multiplier, err := strconv.Atoi(num)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", num)
	}

	var interval int
	switch unit {
	case 'm':
		interval = multiplier * 60
	case 'h':
		interval = multiplier * 60 * 60
	case 'd':
		interval = multiplier * 24 * 60 * 60
	case 'w':
		interval = multiplier * 7 * 24 * 60 * 60
	default:
		return 0, fmt.Errorf("unsupported unit type: %c", unit)
	}

	return interval, nil
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func RunCmd(c string, param ...string) (string, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command(c, param...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	} else {
		return out.String(), nil
	}
}
