/*
Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved
*/

package postgresql

import "gorm.io/gorm"

// Impl is the interface for the DAO.
type Impl interface {
	DB() *gorm.DB
	TableName() string
}

// DAO creates a new daoImpl instance with the specified table name.
func DAO(table string) *daoImpl {
	return &daoImpl{
		table: table,
	}
}

type daoImpl struct {
	table string
}

// DB Each operation must generate a new gorm.DB instance.
// If using the same gorm.DB instance by different operations, they will share the same error.
func (dao *daoImpl) DB() *gorm.DB {
	return db.Table(dao.table)
}

// TableName returns the name of the table associated with this daoImpl instance.
func (dao *daoImpl) TableName() string {
	return dao.table
}
