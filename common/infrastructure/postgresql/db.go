/*
Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved
*/

// Package postgresql provides functionality for interacting with PostgreSQL databases.
package postgresql

import (
	"errors"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

var (
	db         *gorm.DB
	errorCodes errorCode
)

var serverLogger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	logger.Config{
		LogLevel:                  logger.Warn, // Log level
		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
		ParameterizedQueries:      true,        // Don't include params in the SQL log
		Colorful:                  false,       // Disable color
	},
)

// Init initializes the database connection and configuration.
func Init(cfg *Config, removeCfg bool) error {
	defer cfg.clear()

	dbInstance, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN: cfg.dsn(),
			// disables implicit prepared statement usage
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{
			Logger:                 serverLogger,
			SkipDefaultTransaction: true,
		},
	)
	if err != nil {
		return err
	}

	if err := dbInstance.Use(tracing.NewPlugin()); err != nil {
		panic(err)
	}

	if removeCfg && cfg.Dbcert != "" {
		if err := os.Remove(cfg.Dbcert); err != nil {
			return err
		}
	}

	sqlDb, err := dbInstance.DB()
	if err != nil {
		return err
	}

	sqlDb.SetConnMaxLifetime(cfg.getLifeDuration())
	sqlDb.SetMaxOpenConns(cfg.MaxConn)
	sqlDb.SetMaxIdleConns(cfg.MaxIdle)

	db = dbInstance

	errorCodes = cfg.Code

	return nil
}

// DB returns the current database instance.
func DB() *gorm.DB {
	return db
}

// AutoMigrate automatically migrates the given table.
func AutoMigrate(table interface{}) error {
	// pointer non-nil check
	if db == nil {
		err := errors.New("empty pointer of *gorm.DB")
		logrus.Error(err.Error())

		return err
	}

	return db.AutoMigrate(table)
}
