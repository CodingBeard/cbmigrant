package model

import (
	"fmt"
	"time"
)

type MigrantMigration struct {
	Database string `gorm:"-"`
	Table    string `gorm:"-"`

	Id       int
	Name     string
	Migrated time.Time
	Batch    int
}

func (m *MigrantMigration) TableName() string {
	return fmt.Sprintf("`%s`.`%s`", m.Database, m.Table)
}
