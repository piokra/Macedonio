package macedonio

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
)

var dbHandle *gorm.DB

func GetDBHandle() *gorm.DB {
	return dbHandle
}

func InitDBHandle() error {
	if dbHandle != nil {
		return fmt.Errorf("handle reinitialization")
	}

	host := os.Getenv("MYSQL_HOSTNAME")
	database := os.Getenv("MYSQL_DB")
	user := os.Getenv("MYSQL_USER")
	pwd := os.Getenv("MYSQL_PASSWORD")

	db, err := gorm.Open("mysql", fmt.Sprintf("Server=%s;Database=%s;Uid=%s;Pwd=%s", host, database, user, pwd))
	if err != nil {
		return err
	}

	dbHandle = db
	return nil
}
