package task

import (
	"errors"
	"fmt"
	"github.com/codingbeard/cbmigrant"
	"github.com/codingbeard/cbmigrant/model"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

type MigrantMigrate struct {
	Db                        *gorm.DB
	CreateMigrantDatabaseFunc func(db *gorm.DB) error
	MigrantDatabaseName       string
	MigrantTableName          string
	Logger                    cbmigrant.Logger
	ErrorHandler              cbmigrant.ErrorHandler
}

func (m MigrantMigrate) GetSchedule() string {
	return "manual"
}

func (m MigrantMigrate) GetGroup() string {
	return "migrant"
}

func (m MigrantMigrate) GetName() string {
	return "migrate"
}

func (m MigrantMigrate) Run() error {
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

	e := m.Db.AutoMigrate(&model.MigrantMigration{}).Error
	if e != nil {
		m.ErrorHandler.Error(e)
		return e
	}

	var dbMigrations []*model.MigrantMigration
	e = m.Db.Find(&dbMigrations).Error
	if e != nil {
		m.ErrorHandler.Error(e)
		return e
	}

	batch := 0
	for _, mig := range dbMigrations {
		if mig.Batch > batch {
			batch = mig.Batch
		}
	}

	batch++

	for _, mig := range cbmigrant.GetMigrations() {
		found := false
		for _, dbMigration := range dbMigrations {
			if dbMigration.Name == mig.Name {
				found = true
			}
		}

		if !found {
			m.Logger.InfoF("MIGRANT", "Running migration %s", mig.Name)
			e := mig.Up(m.Db)
			if e != nil {
				m.ErrorHandler.Error(e)
				return e
			}
			m.Db.Create(&model.MigrantMigration{
				Name:     mig.Name,
				Migrated: time.Now(),
				Batch:    batch,
			})
			m.Logger.InfoF("MIGRANT", "Finished migration %s", mig.Name)
		}
	}

	return nil
}
