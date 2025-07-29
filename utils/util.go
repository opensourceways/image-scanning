package utils

import (
	"fmt"
	"os"
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
