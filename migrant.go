package cbmigrant

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"runtime"
	"sort"
	"strings"
)

type Migration struct {
	Name string
	Up   func(db *gorm.DB) error
	Down func(db *gorm.DB) error
}

var Migrations []Migration

func GetMigrations() []Migration {
	sort.Slice(Migrations, func(i, j int) bool {
		return Migrations[i].Name < Migrations[j].Name
	})
	return Migrations
}

type Logger interface {
	InfoF(category string, message string, args ...interface{})
}

type DefaultLogger struct{}

func (d DefaultLogger) InfoF(category string, message string, args ...interface{}) {
	log.Println(category+":", fmt.Sprintf(message, args...))
}

type ErrorHandler interface {
	Error(e error)
}

type DefaultErrorHandler struct{}

func (d DefaultErrorHandler) Error(e error) {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	stack := string(buf)
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	log.Println("ERROR", e.Error()+"\n"+stack)
}