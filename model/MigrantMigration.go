package model

import (
	"time"
)

type MigrantMigration struct {
	Id       int
	Name     string
	Migrated time.Time
	Batch    int
}
