package migranttask

import (
	"github.com/codingbeard/cbmigrant"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type MigrantMake struct {
	FolderPath   string
	PackageName  string
	Logger       cbmigrant.Logger
	ErrorHandler cbmigrant.ErrorHandler
}

func (m MigrantMake) GetSchedule() string {
	return "manual"
}

func (m MigrantMake) GetGroup() string {
	return "migrant"
}

func (m MigrantMake) GetName() string {
	return "make"
}

func (m MigrantMake) Run() error {
	if m.Logger == nil {
		m.Logger = cbmigrant.DefaultLogger{}
	}
	if m.ErrorHandler == nil {
		m.ErrorHandler = cbmigrant.DefaultErrorHandler{}
	}
	if m.FolderPath == "" {
		m.FolderPath = "migration/"
	}
	if m.PackageName == "" {
		m.PackageName = "migration"
	}
	if len(os.Args) > 5 && os.Args != nil {
		database := os.Args[3]
		table := os.Args[4]
		name := os.Args[5]
		date := time.Now()
		datetime := date.Format("2006_01_02_150405_")

		migrationName := datetime + name
		fileName := m.FolderPath + migrationName + ".go"
		content := stringReplacer(`package {packageName}

import (
	"github.com/codingbeard/cbmigrant"
	"github.com/jinzhu/gorm"
)

func init() {
	cbmigrant.Migrations = append(cbmigrant.Migrations, cbmigrant.Migration{
		Name: "{migrationName}",
		Up: func(db *gorm.DB) error {
			return db.Table("{database}.{table}").Error
		},
		Down: func(db *gorm.DB) error {
			return db.Table("{database}.{table}").Error
		},
	})
}`, "{packageName}", m.PackageName,
			"{migrationName}", migrationName,
			"{database}", database,
			"{table}", table)

		e := ioutil.WriteFile(fileName, []byte(content), os.ModePerm)
		if e != nil {
			m.ErrorHandler.Error(e)
		}

		m.Logger.InfoF("TASK", "Created a blank migration %s", fileName)
	} else {
		m.Logger.InfoF("TASK", "Not enough arguments given, expecting database, table, name")
	}

	return nil
}

func stringReplacer(format string, args ...string) string {
	r := strings.NewReplacer(args...)
	return r.Replace(format)
}
