package mysql

//mysql数据库连接
import (
	"database/sql"
	"fmt"
	"github.com/chenxuan520/goweb-platform/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _defaultDB *gorm.DB

func Init(dsn string) (*gorm.DB, error) {

	mysqlConfig := mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		SkipInitializeWithVersion: false,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{}); err != nil {
		return nil, err
	} else {
		_defaultDB = db
		return db, nil
	}
}

func CreateMysqlDsn(username, password, path, port, dbname, config string) string {
	return username + ":" + password + "@tcp(" + path + ":" + port + ")/" + dbname + "?" + config
}

func GetMysqlDB() *gorm.DB {
	if _defaultDB == nil {
		logger.GetLogger().Error("mysql database is not initialized")
		return nil
	}
	return _defaultDB
}

//to create database
func CreateDatabase(dsn string, driver string, createSql string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(db)
	if err = db.Ping(); err != nil {
		return err
	}
	_, err = db.Exec(createSql)
	return err
}
