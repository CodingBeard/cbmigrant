package migranttask

import (
	"errors"
	"fmt"
	"github.com/codingbeard/cbmigrant"
	"github.com/codingbeard/cbmigrant/model"
	"github.com/jinzhu/gorm"
	"strings"
)

type MigrantRollback struct {
	Db                        *gorm.DB
	CreateMigrantDatabaseFunc func(db *gorm.DB) error
	MigrantDatabaseName       string
	MigrantTableName          string
	Logger                    cbmigrant.Logger
	ErrorHandler              cbmigrant.ErrorHandler
}

func (m MigrantRollback) GetSchedule() string {
	return "manual"
}

func (m MigrantRollback) GetGroup() string {
	return "migrant"
}

func (m MigrantRollback) GetName() string {
	return "rollback"
}

func (m MigrantRollback) Run() error {
	if m.Logger == nil {
		m.Logger = cbmigrant.DefaultLogger{}
	}
	if m.ErrorHandler == nil {
		m.ErrorHandler = cbmigrant.DefaultErrorHandler{}
	}
	if m.Db == nil {
		e := errors.New("no database set on task")
		m.ErrorHandler.Error(e)
		return e
	}
	if m.MigrantDatabaseName == "" {
		m.MigrantDatabaseName = "migrant"
	}
	if m.MigrantTableName == "" {
		m.MigrantTableName = "migration"
	}
	if m.MigrantDatabaseName == "" || m.MigrantTableName == "" {
		e := errors.New("migrant database or table name not configured")
		m.ErrorHandler.Error(e)
		return e
	}

	if m.CreateMigrantDatabaseFunc == nil {
		e := m.Db.Exec(fmt.Sprintf("USE `%s`", m.MigrantDatabaseName)).Error
		if e != nil {
			if strings.Contains(e.Error(), fmt.Sprintf("Unknown database '%s'", m.MigrantDatabaseName)) {
				m.Logger.InfoF("MIGRANT", "Creating database discord")
				e := m.Db.Exec(fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", m.MigrantDatabaseName)).Error
				if e != nil {
					m.ErrorHandler.Error(e)
					return e
				}
			} else {
				m.ErrorHandler.Error(e)
				return e
			}
		}
	} else {
		e := m.CreateMigrantDatabaseFunc(m.Db)
		if e != nil {
			m.ErrorHandler.Error(e)
			return e
		}
	}

	e := m.Db.Table(m.MigrantDatabaseName+"."+m.MigrantTableName).AutoMigrate(&model.MigrantMigration{}).Error
	if e != nil {
		m.ErrorHandler.Error(e)
		return e
	}

	var dbMigrations []*model.MigrantMigration
	e = m.Db.Table(m.MigrantDatabaseName+"."+m.MigrantTableName).Find(&dbMigrations).Error
	if e != nil {
		m.ErrorHandler.Error(e)
		return e
	}

	var reverseDbMigrations []*model.MigrantMigration
	for i := len(dbMigrations) - 1; i >= 0; i-- {
		reverseDbMigrations = append(reverseDbMigrations, dbMigrations[i])
	}

	dbMigrations = reverseDbMigrations

	batch := 0
	for _, mig := range dbMigrations {
		if mig.Batch > batch {
			batch = mig.Batch
		}
	}

	for _, dbMigration := range reverseDbMigrations {
		for _, mig := range cbmigrant.GetMigrations() {
			if mig.Name == dbMigration.Name && dbMigration.Batch == batch {
				m.Logger.InfoF("TASK", "Rolling back migration %s", mig.Name)
				e := mig.Down(m.Db)
				if e != nil {
					m.ErrorHandler.Error(e)
				}
				m.Db.Table(m.MigrantDatabaseName+"."+m.MigrantTableName).Delete(dbMigration)
				m.Logger.InfoF("TASK", "Rolled back migration %s", mig.Name)
			}
		}
	}

	return nil
}
