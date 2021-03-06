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
	pwd := os.Getenv("MYSQL_PASSWORD")
	db, err := gorm.Open("mysql", fmt.Sprintf("macedonio:%s@(%s)/Macedonio?charset=utf8&parseTime=True&loc=Local", pwd, host))
	if err != nil {
		return err
	}

	dbHandle = db
	return nil
}
