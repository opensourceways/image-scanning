/*
Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved
*/

// Package postgresql provides functionality for interacting with PostgreSQL databases.
package postgresql

import (
	"fmt"
	"time"
)

// Config represents the configuration for PostgreSQL.
type Config struct {
	Host    string    `json:"host"     required:"true"`
	User    string    `json:"user"     required:"true"`
	Pwd     string    `json:"pwd"      required:"true"`
	Name    string    `json:"name"     required:"true"`
	Port    int       `json:"port"     required:"true"`
	Life    int       `json:"life"     required:"true"` // the unit is minute
	MaxConn int       `json:"max_conn" required:"true"`
	MaxIdle int       `json:"max_idle" required:"true"`
	Dbcert  string    `json:"cert"`
	Code    errorCode `json:"error_code"`
}

// SetDefault sets the default values for the Config.
func (cfg *Config) SetDefault() {
	if cfg.MaxConn <= 0 {
		cfg.MaxConn = 500
	}

	if cfg.MaxIdle <= 0 {
		cfg.MaxIdle = 250
	}

	if cfg.Life <= 0 {
		cfg.Life = 2
	}
}

// ConfigItems returns the configuration items for the Config.
func (cfg *Config) ConfigItems() []interface{} {
	return []interface{}{
		&cfg.Code,
	}
}

func (cfg *Config) getLifeDuration() time.Duration {
	return time.Minute * time.Duration(cfg.Life)
}

func (cfg *Config) dsn() string {
	if cfg.Dbcert != "" {
		return fmt.Sprintf(
			"host=%v user=%v password=%v dbname=%v port=%v sslmode=verify-ca TimeZone=Asia/Shanghai sslrootcert=%v",
			cfg.Host, cfg.User, cfg.Pwd, cfg.Name, cfg.Port, cfg.Dbcert,
		)
	} else {
		return fmt.Sprintf(
			"host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.User, cfg.Pwd, cfg.Name, cfg.Port,
		)
	}
}

func (cfg *Config) clear() {
	cfg.Host = ""
	cfg.User = ""
	cfg.Pwd = ""
	cfg.Name = ""
}

type errorCode struct {
	UniqueConstraint string `json:"unique_constraint"`
}

// SetDefault sets the default values for the errorCode.
func (cfg *errorCode) SetDefault() {
	if cfg.UniqueConstraint == "" {
		cfg.UniqueConstraint = "23505"
	}
}
