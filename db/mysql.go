package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/config"
	"os"
	"time"
	"v2ray.com/core/common/errors"
)

func NewMySQLConn(config *config.MySQLConfig) (*gorm.DB, error) {
	newError("Connecting database...").AtInfo().WriteToLog()
	defer newError("Connected").AtInfo().WriteToLog()

	dsn, err := config.FormatDSN()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SingularTable(true)

	return db, nil
}

func newErrorf(format string, a ...interface{}) *errors.Error {
	return newError(fmt.Sprintf(format, a...))
}

func newError(values ...interface{}) *errors.Error {
	values = append([]interface{}{"SSPanelPlugin: "}, values...)
	return errors.New(values...)
}

func fatal(values ...interface{}) {
	newError(values...).AtError().WriteToLog()
	// Wait log
	time.Sleep(1 * time.Second)
	os.Exit(-2)
}
